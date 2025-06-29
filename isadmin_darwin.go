package main

import (
	"os/exec"
	"strings"
)

func check_is_admin() (bool, error) {
	// On macOS, check if the current user is in the admin group
	cmd := exec.Command("groups")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Check if "admin" is in the groups output
	groups := string(output)
	return strings.Contains(groups, "admin"), nil
}
