package main

import (
	"fmt"
	"os"
	"os/exec"
)

func UpdateJellyfin() {
	packageName := "jellyfin"

	fmt.Println("Updating package list...")
	cmd := exec.Command("sudo", "apt", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ExitWithError("Failed to update package list", err)
	}

	fmt.Printf("Upgrading package %s...\n", packageName)
	cmd = exec.Command("sudo", "apt", "install", "--only-upgrade", "-y", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ExitWithError("Failed to upgrade Jellyfin package", err)
	}

	fmt.Println("Jellyfin updated successfully.")
}
