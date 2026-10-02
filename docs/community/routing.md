# Routing and settings

Open the application's Tor and WireGuard page for setup steps. Enable **Split Tunnel Module** (`splittun/enable`), choose an application, then enable **Use Split Tunnel** (`splittun/use`) and set **Network Interface** (`splittun/networkInterface`) in that application's settings. Avoid enabling routing globally until exclusions and failure behavior have been tested. Keep the proxy or VPN client's own process outside its route.

## Tor

Run a local Tor daemon. Set the application's Network Interface to `tor` for `127.0.0.1:9050`, or `tor://127.0.0.1:9150` for an explicit SOCKS port. Only literal loopback IPs and valid ports are accepted. Selected TCP connections use the SOCKS proxy; selected UDP connections are rejected. A stalled SOCKS handshake times out, and proxy failure has no direct TCP fallback. Selected unsupported outbound protocols such as ICMP are blocked after routing exclusions are evaluated. System DNS queries remain outside this proxy path.

DNS uses Portmaster's configured resolver separately. Onion names, circuit isolation, and Tor Browser fingerprint protection are not implemented. Other applications do not inherit Tor Browser protections by using this proxy. History remains local and can contain sensitive destinations.

## WireGuard

Use your own WireGuard client, keys, peers, routes, and AllowedIPs. Start the tunnel, then set Network Interface to its interface name or assigned local tunnel IP. Windows binds using the selected interface index; Linux uses its interface-binding mechanism. An unavailable interface fails binding. An interface that still exists while its tunnel is down does not establish a verified kill switch: operating-system routing and peer configuration determine behavior. No cross-platform leak-proof claim is made.

## Verify and troubleshoot

1. Generate traffic from only the selected application and check the application connection view.
2. Compare that application's public IPv4 and IPv6 paths with the tunnel on and off.
3. Stop Tor and confirm selected TCP connections fail; verify UDP is blocked for that application.
4. Stop WireGuard and check traffic does not use an unintended path. Record interface presence, routes, and AllowedIPs.
5. Review DNS resolver configuration and verify DNS separately. Do not assume it follows the application's route.
6. Confirm other applications, the VPN client, and the proxy daemon still work as intended.

Record OS, interface, versions, results, and redacted logs in a compatibility issue. Never upload private keys or full browsing logs. Live end-to-end Tor and WireGuard verification is still an outstanding release requirement.
