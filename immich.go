package main

import (
	"fmt"
	"os"
	"os/exec"
)

func UpdateImmich() {
	fmt.Println("Updating Immich...")

	cmds := [][]string{
		{"docker", "compose", "down"},
		{"docker", "compose", "pull"},
		{"docker", "compose", "up", "-d"},
		{"docker", "image", "prune", "-f"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = ImmichDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			ExitWithError(fmt.Sprintf("command %v failed", args), err)
		}
	}

	fmt.Println("Immich update completed successfully.")
}
