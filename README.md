# WSSHTunnel

A WebSocket proxy implementation for SSH in Go. This tool enables tunneling SSH connections through the WebSocket protocol.

## Features

- SSH connection proxying via WebSocket
- Optional password authentication
- Flexible address and port configuration
- Timeout support for inactive connections
- Simple system installation with systemd
- Multi-connection management using goroutines
- Compatible with standard SSH clients with proper configuration

## Requirements

- Go 1.19 or newer
- System with systemd support (for service installation)
- Administrative privileges (for service installation)

## Installation

### Installation from Source

Clone this repository and build from source:

```bash
git clone https://github.com/sauufi/wssht.git
cd wssht
./scripts/build.sh
```

Binary will be created in the `bin/` directory.

### Installation as a Service (Systemd)

For complete installation as a systemd service, run:

```bash
sudo ./scripts/install.sh
```

This script will:

1. Build the binary if it doesn't exist
2. Install the binary to `/usr/local/bin/`
3. Create a configuration file in `/etc/wssht/`
4. Install and enable the systemd service
5. Start the service

## Usage

### Command Line

Run the proxy with default options:

```bash
wssht
```

Specify address, port and default target host:

```bash
wssht -b 0.0.0.0 -p 8080 -t 127.0.0.1:22
```

Use password for authentication:

```bash
wssht -b 0.0.0.0 -p 8080 -t 127.0.0.1:22 -pass your_secure_password
```

### Service Management

Start the service:

```bash
sudo systemctl start wssht.service
```

Stop the service:

```bash
sudo systemctl stop wssht.service
```

View logs:

```bash
sudo journalctl -u wssht.service -f
```

### Configuration

If installed as a service, edit the configuration file:

```bash
sudo nano /etc/wssht/config
```

Then restart the service:

```bash
sudo systemctl restart wssht.service
```

## How It Works

The proxy accepts WebSocket connections and creates a tunnel to the target SSH server. Clients send an `X-Real-Host` header to specify the destination host. Optional authentication can be implemented with the `X-Pass` header.

WebSocket protocol is used to traverse firewalls or proxies that might block direct SSH traffic.

## Security

- Be sure to use a strong password if enabling the authentication feature
- Consider using a reverse proxy with HTTPS for secure connections
- Restrict access to the proxy server if possible
- Monitor logs for suspicious activity

## Code Structure

```
wssht/
├── cmd/wssht/         # Main application code
├── internal/tunnel/   # Internal packages for server and handler
├── scripts/           # Build and installation scripts
├── systemd/           # Systemd configuration files
└── README.md          # Documentation
```

## Contributing

Contributions to this project are welcome. Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the [MIT License](LICENSE).
