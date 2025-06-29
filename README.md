# Tailscale Tools

A Go application that automates the setup of Tailscale funnel tunnels for WordPress sites. This tool simplifies the process of creating temporary public URLs for local WordPress development environments.

## Features

- **Automatic Funnel Setup**: Creates Tailscale funnel tunnels with a single command
- **WordPress Integration**: Automatically updates WordPress configuration files
- **Apache Integration**: Updates Apache virtual host configurations (Windows/WAMP)
- **Cross-Platform Support**: Works on Windows, Linux, and macOS
- **Automatic Cleanup**: Reverts all changes when the tunnel is closed
- **Admin Privilege Check**: Ensures proper permissions before making system changes

## Prerequisites

- [Go](https://golang.org/dl/) 1.24.0 or later
- [Tailscale](https://tailscale.com/) installed and authenticated
- WordPress site running locally
- Apache (for Windows/WAMP integration)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/shujink0/tailscale-tools.git
cd tailscale-tools

# Build the application
go build -o tailscale-tools
```

### Cross-Platform Builds

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o tailscale-tools.exe

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o tailscale-tools

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o tailscale-tools
```

## Usage

### Basic Usage

```bash
# Start a funnel tunnel on port 80 for a WordPress site
tailscale-tools --port 80 --host yoursite.local

# Use a different port
tailscale-tools --port 8080 --host yoursite.local

# Specify Apache version (Windows/WAMP only)
tailscale-tools --port 80 --host yoursite.local --apachev 2.4.62.1
```

### Command Line Options

- `--port`: Port number to forward (default: "80")
- `--host`: Hostname to be replaced in configuration files (default: "ferreteria.cifu.dev")
- `--apachev`: Apache HTTPD version for WAMP (default: "2.4.54.2")

### Examples

#### Windows/WAMP Setup

```bash
# For a WordPress site at C:\wamp64\www\mysite\
tailscale-tools.exe --port 80 --host mysite.local --apachev 2.4.62.1
```

#### Linux Setup

```bash
# For a WordPress site at /var/www/mysite/
tailscale-tools --port 80 --host mysite.local
```

#### macOS Setup

```bash
# For a WordPress site
tailscale-tools --port 80 --host mysite.local
```

## How It Works

1. **Privilege Check**: Verifies that the application is running with admin privileges
2. **Funnel Creation**: Starts a Tailscale funnel tunnel on the specified port
3. **Configuration Update**: 
   - Updates WordPress `wp-config.php` file
   - Updates Apache virtual hosts configuration (Windows only)
4. **Service Restart**: Restarts Apache service (Windows only)
5. **User Interaction**: Waits for user input to close the tunnel
6. **Cleanup**: Reverts all configuration changes and stops the funnel

## File Paths

The application automatically determines file paths based on the operating system:

### Windows
- WordPress config: `C:\wamp64\www\{subdomain}\wp-config.php`
- Apache vhosts: `C:\wamp64\bin\apache\apache{version}\conf\extra\httpd-vhosts.conf`

### Linux
- WordPress config: `/var/www/{subdomain}/wp-config.php`

## Security Considerations

- The application requires admin privileges to modify system files
- File paths are validated to prevent path traversal attacks
- All changes are automatically reverted when the tunnel is closed
- The application only modifies specific configuration files

## Troubleshooting

### Common Issues

1. **"Elevated privileges required"**: Run the application as Administrator (Windows) or with sudo (Linux/macOS)

2. **"No funnel URL found"**: Ensure Tailscale is properly authenticated and running

3. **"File not found"**: Verify that the WordPress site exists at the expected path

4. **Apache restart fails**: Ensure the WAMP Apache service is properly installed and configured

### Debug Mode

For troubleshooting, you can add verbose logging by modifying the source code or checking the console output for detailed execution steps.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Dependencies

- `golang.org/x/sys`: For Windows admin privilege checking
- `mvdan.cc/xurls`: For URL extraction from command output
