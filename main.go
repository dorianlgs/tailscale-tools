package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"mvdan.cc/xurls"
)

// Config holds the application configuration
type Config struct {
	Port          string
	Host          string
	ApacheVersion string
}

// App represents the main application
type App struct {
	config *Config
}

// NewApp creates a new application instance
func NewApp(config *Config) *App {
	return &App{config: config}
}

func main() {
	config, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	app := NewApp(config)

	if err := app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() (*Config, error) {
	port := flag.String("port", "80", "port to forward")
	host := flag.String("host", "ferreteria.cifu.dev", "host to be replaced")
	apacheVersion := flag.String("apachev", "2.4.54.2", "apache httpd version of wamp")

	flag.Parse()

	if *port == "" {
		return nil, fmt.Errorf("port parameter required")
	}

	if *host == "" {
		return nil, fmt.Errorf("host parameter required")
	}

	subDomain := strings.Split(*host, ".")
	if len(subDomain) == 0 {
		return nil, fmt.Errorf("did not find subdomain in host parameter")
	}

	return &Config{
		Port:          *port,
		Host:          *host,
		ApacheVersion: *apacheVersion,
	}, nil
}

func (app *App) run() error {
	// Check admin privileges
	isAdmin, err := check_is_admin()
	if err != nil {
		return fmt.Errorf("failed to check admin privileges: %w", err)
	}

	if !isAdmin {
		return fmt.Errorf("elevated privileges required, please run as Administrator")
	}

	// Get file paths
	wpConfigPath, vHostsPath, err := app.getFilePaths()
	if err != nil {
		return fmt.Errorf("failed to determine file paths: %w", err)
	}

	// Start Tailscale funnel
	funnelURL, err := app.startFunnel()
	if err != nil {
		return fmt.Errorf("failed to start funnel: %w", err)
	}

	// Update configuration files
	if err := app.updateConfigFiles(wpConfigPath, vHostsPath, funnelURL.Host); err != nil {
		return fmt.Errorf("failed to update config files: %w", err)
	}

	// Restart Apache if on Windows
	if runtime.GOOS == "windows" {
		if err := app.restartApache(); err != nil {
			return fmt.Errorf("failed to restart Apache: %w", err)
		}
	}

	fmt.Printf("Share the URL: %s\n", funnelURL.String())
	fmt.Println()
	fmt.Println("Press Enter to close the tunnel and revert config changes (Apache and WordPress)")

	// Wait for user input
	reader := bufio.NewReader(os.Stdin)
	if _, err := reader.ReadString('\n'); err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	// Cleanup
	return app.cleanup(wpConfigPath, vHostsPath, funnelURL.Host)
}

func (app *App) getFilePaths() (wpConfigPath, vHostsPath string, err error) {
	subDomain := strings.Split(app.config.Host, ".")
	if len(subDomain) == 0 {
		return "", "", fmt.Errorf("invalid host format")
	}

	switch runtime.GOOS {
	case "windows":
		wpConfigPath = fmt.Sprintf("C:\\wamp64\\www\\%s\\wp-config.php", subDomain[0])
		vHostsPath = fmt.Sprintf("C:\\wamp64\\bin\\apache\\apache%s\\conf\\extra\\httpd-vhosts.conf", app.config.ApacheVersion)
	case "linux":
		wpConfigPath = fmt.Sprintf("/var/www/%s/wp-config.php", subDomain[0])
		vHostsPath = ""
	default:
		return "", "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return wpConfigPath, vHostsPath, nil
}

func (app *App) startFunnel() (*url.URL, error) {
	fmt.Println("Starting Tailscale funnel...")

	output, err := execute("tailscale", "funnel", "--bg", app.config.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to start funnel: %w", err)
	}

	rxRelaxed := xurls.Relaxed
	funnelURLString := rxRelaxed.FindString(output)
	if funnelURLString == "" {
		return nil, fmt.Errorf("no funnel URL found in output")
	}

	funnelURL, err := url.Parse(funnelURLString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse funnel URL: %w", err)
	}

	return funnelURL, nil
}

func (app *App) updateConfigFiles(wpConfigPath, vHostsPath, funnelHost string) error {
	// Create backups before modifying files
	if err := app.createBackup(wpConfigPath); err != nil {
		return fmt.Errorf("failed to create WordPress config backup: %w", err)
	}

	// Update WordPress config
	if err := replaceTextInFile(wpConfigPath, app.config.Host, funnelHost); err != nil {
		return fmt.Errorf("failed to update WordPress config: %w", err)
	}

	// Update vhosts config on Windows
	if runtime.GOOS == "windows" && vHostsPath != "" {
		if err := app.createBackup(vHostsPath); err != nil {
			return fmt.Errorf("failed to create vhosts config backup: %w", err)
		}

		if err := replaceTextInFile(vHostsPath, app.config.Host, funnelHost); err != nil {
			return fmt.Errorf("failed to update vhosts config: %w", err)
		}
	}

	return nil
}

func (app *App) createBackup(filePath string) error {
	if filePath == "" {
		return nil
	}

	backupPath := filePath + ".backup"

	// Check if backup already exists
	if _, err := os.Stat(backupPath); err == nil {
		fmt.Printf("Backup already exists: %s\n", backupPath)
		return nil
	}

	// Read original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	// Create backup
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	fmt.Printf("Created backup: %s\n", backupPath)
	return nil
}

func (app *App) restartApache() error {
	fmt.Println("Restarting Apache...")

	if _, err := execute("net", "stop", "wampapache64"); err != nil {
		return fmt.Errorf("failed to stop Apache: %w", err)
	}

	if _, err := execute("net", "start", "wampapache64"); err != nil {
		return fmt.Errorf("failed to start Apache: %w", err)
	}

	return nil
}

func (app *App) cleanup(wpConfigPath, vHostsPath, funnelHost string) error {
	fmt.Println("Cleaning up...")

	// Reset funnel
	if _, err := execute("tailscale", "funnel", "reset"); err != nil {
		return fmt.Errorf("failed to reset funnel: %w", err)
	}

	// Restore WordPress config from backup if available
	if err := app.restoreFromBackup(wpConfigPath); err != nil {
		// Fallback to manual replacement if backup restore fails
		if err := replaceTextInFile(wpConfigPath, funnelHost, app.config.Host); err != nil {
			return fmt.Errorf("failed to revert WordPress config: %w", err)
		}
	}

	// Restore vhosts config on Windows
	if runtime.GOOS == "windows" && vHostsPath != "" {
		if err := app.restoreFromBackup(vHostsPath); err != nil {
			// Fallback to manual replacement if backup restore fails
			if err := replaceTextInFile(vHostsPath, funnelHost, app.config.Host); err != nil {
				return fmt.Errorf("failed to revert vhosts config: %w", err)
			}
		}
	}

	// Restart Apache if on Windows
	if runtime.GOOS == "windows" {
		if err := app.restartApache(); err != nil {
			return fmt.Errorf("failed to restart Apache during cleanup: %w", err)
		}
	}

	fmt.Println("Cleanup completed successfully")
	return nil
}

func (app *App) restoreFromBackup(filePath string) error {
	if filePath == "" {
		return nil
	}

	backupPath := filePath + ".backup"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupPath)
	}

	// Read backup file
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Restore original file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	fmt.Printf("Restored from backup: %s\n", backupPath)
	return nil
}

func execute(program string, args ...string) (string, error) {
	fmt.Printf("Executing %s with args: %s\n", program, strings.Join(args, " "))

	cmd := exec.Command(program, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	fmt.Println(string(output))
	return string(output), nil
}

func replaceTextInFile(filePath, find, replace string) error {
	// Validate file path to prevent path traversal
	if !filepath.IsAbs(filePath) {
		return fmt.Errorf("file path must be absolute: %s", filePath)
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	fileName := filepath.Base(filePath)
	fmt.Printf("Replacing in file %s: '%s' -> '%s'\n", fileName, find, replace)

	// Replace text
	newContent := strings.ReplaceAll(string(content), find, replace)

	// Write file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
