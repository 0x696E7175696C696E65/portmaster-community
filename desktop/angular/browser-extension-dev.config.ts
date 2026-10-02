import type { Configuration } from 'webpack';
const config = require('./browser-extension.config');

// Reload the extension manually after rebuilding. The retired hot-reloader
// depended on an unmaintained vulnerable user-agent parser.
module.exports = { ...config, mode: 'development' } as Configuration;
