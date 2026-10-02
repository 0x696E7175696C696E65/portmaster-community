const { test } = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const http = require('node:http');
const { once } = require('node:events');
const Piscina = require('piscina');
const sockjs = require('sockjs');
const WebSocket = require('ws');

test('worker pools ignore inherited execution options after prototype pollution', { timeout: 10000 }, async () => {
  let pool;
  Object.prototype.execArgv = ['--require', path.join(__dirname, 'nonexistent-injected-module.cjs')];
  try {
    pool = new Piscina({ filename: path.join(__dirname, 'fixtures/pool-worker.cjs'), minThreads: 1, maxThreads: 1 });
  } finally {
    delete Object.prototype.execArgv;
  }
  try {
    assert.equal(await pool.run(1), 2);
  } finally {
    await pool.destroy();
  }
});

test('patched UUID dependency preserves local SockJS session and message handling', { timeout: 10000 }, async () => {
  const transport = sockjs.createServer({ log: () => {} });
  transport.on('connection', connection => connection.on('data', data => connection.write(data)));
  const server = http.createServer();
  const connections = new Set();
  server.on('connection', connection => {
    connections.add(connection);
    connection.on('close', () => connections.delete(connection));
  });
  transport.installHandlers(server, { prefix: '/test' });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  const socket = new WebSocket(`ws://127.0.0.1:${server.address().port}/test/000/security/websocket`);
  const frames = [];
  const waiters = [];
  socket.on('message', frame => {
    const waiter = waiters.shift();
    if (waiter) waiter(frame);
    else frames.push(frame);
  });
  const nextFrame = () => frames.length ? Promise.resolve(frames.shift()) : new Promise(resolve => waiters.push(resolve));
  try {
    await once(socket, 'open');
    const opened = await nextFrame();
    assert.equal(opened.toString(), 'o');
    socket.send(JSON.stringify(['security echo']));
    const echoed = await nextFrame();
    assert.equal(echoed.toString(), 'a["security echo"]');
  } finally {
    socket.terminate();
    for (const connection of connections) connection.destroy();
    await new Promise(resolve => server.close(resolve));
  }
});
