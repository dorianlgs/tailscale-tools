package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseFlags(t *testing.T) {
	// Test with valid flags
	os.Args = []string{"tailscale-tools", "--port", "8080", "--host", "test.local"}

	config, err := parseFlags()
	if err != nil {
		t.Fatalf("parseFlags() failed: %v", err)
	}

	if config.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", config.Port)
	}

	if config.Host != "test.local" {
		t.Errorf("Expected host test.local, got %s", config.Host)
	}
}

func TestGetFilePaths(t *testing.T) {
	app := &App{
		config: &Config{
			Host:          "test.local",
			ApacheVersion: "2.4.54.2",
		},
	}

	wpConfigPath, vHostsPath, err := app.getFilePaths()

	// On unsupported OS, we expect an error
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		if err == nil {
			t.Error("Expected error for unsupported OS")
		}
		return
	}

	if err != nil {
		t.Fatalf("getFilePaths() failed: %v", err)
	}

	// Check that paths are not empty
	if wpConfigPath == "" {
		t.Error("WordPress config path should not be empty")
	}

	// Check that paths contain the expected subdomain
	if !filepath.IsAbs(wpConfigPath) {
		t.Error("WordPress config path should be absolute")
	}

	// Check vhosts path (may be empty on non-Windows)
	t.Logf("WordPress config path: %s", wpConfigPath)
	t.Logf("VHosts path: %s", vHostsPath)
}

func TestCreateBackup(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"

	// Write test file
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	app := &App{}

	// Create backup
	if err := app.createBackup(testFile); err != nil {
		t.Fatalf("createBackup() failed: %v", err)
	}

	// Check that backup file exists
	backupFile := testFile + ".backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Check that backup content matches original
	backupContent, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != testContent {
		t.Error("Backup content does not match original")
	}
}

func TestReplaceTextInFile(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "old.host.com"

	// Write test file
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Replace text
	if err := replaceTextInFile(testFile, "old.host.com", "new.host.com"); err != nil {
		t.Fatalf("replaceTextInFile() failed: %v", err)
	}

	// Check that content was replaced
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	expected := "new.host.com"
	if string(content) != expected {
		t.Errorf("Expected %s, got %s", expected, string(content))
	}
}
