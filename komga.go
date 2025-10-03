package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// TODO: Download the latest jar file from GitHub releases automatically
func UpdateKomga() {
	latestJarPath, latestVersion, err := findLatestJar()
	if err != nil {
		ExitWithError("Failed to find latest Komga jar", err)
	}

	currentVersion, err := getCurrentVersion()
	if err != nil {
		ExitWithError("Failed to get current Komga version", err)
	}

	fmt.Printf("Current version in service: %s\n", currentVersion)
	fmt.Printf("Latest version available : %s\n", latestVersion)

	if compareSemVer(latestVersion, currentVersion) > 0 {
		fmt.Printf("Updating service file to use version %s\n", latestVersion)
		if err := updateServiceFile(latestVersion); err != nil {
			ExitWithError("Failed to update service file", err)
		}
		if err := reloadAndRestartService(); err != nil {
			ExitWithError("Failed to reload/restart Komga service", err)
		}
		fmt.Printf("Service updated and restarted with version %s\n", latestVersion)

		// Delete old jar files after successful update
		if err := deleteStaleJars(latestJarPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to delete stale jar files: %v\n", err)
		}
	} else {
		fmt.Println("Service already using the latest version. No update needed.")
	}
}

func findLatestJar() (string, string, error) {
	var jars []string
	err := filepath.WalkDir(KomgaDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "komga-") && strings.HasSuffix(d.Name(), ".jar") {
			jars = append(jars, path)
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}
	if len(jars) == 0 {
		return "", "", fmt.Errorf("no komga-*.jar files found in %s", KomgaDir)
	}

	sort.Slice(jars, func(i, j int) bool {
		return compareSemVer(extractVersion(jars[i]), extractVersion(jars[j])) < 0
	})

	latestJar := jars[len(jars)-1]
	return latestJar, extractVersion(latestJar), nil
}

func extractVersion(path string) string {
	base := filepath.Base(path)
	re := regexp.MustCompile(`komga-([0-9.]+)\.jar`)
	matches := re.FindStringSubmatch(base)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func getCurrentVersion() (string, error) {
	file, err := os.Open(KomgaServiceFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`komga-([0-9.]+)\.jar`)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "ExecStart=") {
			matches := re.FindStringSubmatch(line)
			if len(matches) < 2 {
				break
			}
			return matches[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no ExecStart line found in %s", KomgaServiceFile)
}

func updateServiceFile(newVersion string) error {
	input, err := os.ReadFile(KomgaServiceFile)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`komga-[0-9.]+\.jar`)
	updated := re.ReplaceAllString(string(input), fmt.Sprintf("komga-%s.jar", newVersion))
	return os.WriteFile(KomgaServiceFile, []byte(updated), 0644)
}

func reloadAndRestartService() error {
	serviceName := strings.TrimSuffix(filepath.Base(KomgaServiceFile), ".service")
	commands := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "restart", serviceName},
	}
	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %v failed: %w", cmdArgs, err)
		}
	}
	return nil
}

func deleteStaleJars(latestJarPath string) error {
	var deleted int
	err := filepath.WalkDir(KomgaDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "komga-") &&
			strings.HasSuffix(d.Name(), ".jar") && path != latestJarPath {
			fmt.Printf("Deleting old jar: %s\n", path)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to delete %s: %w", path, err)
			}
			deleted++
		}
		return nil
	})
	if err != nil {
		return err
	}
	if deleted > 0 {
		fmt.Printf("Deleted %d stale jar file(s)\n", deleted)
	}
	return nil
}

// compareSemVer compares two semantic versions.
// Returns 1 if v1>v2, -1 if v1<v2, 0 if equal
func compareSemVer(v1, v2 string) int {
	split := func(v string) []int {
		parts := strings.Split(v, ".")
		ints := make([]int, len(parts))
		for i, p := range parts {
			val, _ := strconv.Atoi(p)
			ints[i] = val
		}
		return ints
	}

	a, b := split(v1), split(v2)
	for i := 0; i < len(a) || i < len(b); i++ {
		var ai, bi int
		if i < len(a) {
			ai = a[i]
		}
		if i < len(b) {
			bi = b[i]
		}
		if ai > bi {
			return 1
		} else if ai < bi {
			return -1
		}
	}
	return 0
}
