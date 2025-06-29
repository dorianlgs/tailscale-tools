package main

import (
	"os"
	"os/exec"
	"strings"
)

func check_is_admin() (bool, error) {
	// On Linux, check if the current user is root or in the sudo group
	if os.Geteuid() == 0 {
		return true, nil
	}

	// Check if user is in sudo group
	cmd := exec.Command("groups")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	groups := string(output)
	return strings.Contains(groups, "sudo"), nil
}
