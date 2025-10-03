package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// UpdateKomga checks for updates and downloads if a newer version is available
func UpdateKomga() {
	// Get current installed version
	currentVersion, err := getCurrentVersion()
	if err != nil {
		ExitWithError("Failed to get current Komga version", err)
	}
	fmt.Printf("Current version in service: %s\n", currentVersion)

	// Fetch latest release info from GitHub
	fmt.Println("Checking for updates from GitHub...")
	release, err := fetchLatestRelease()
	if err != nil {
		ExitWithError("Failed to fetch latest release from GitHub", err)
	}

	// Extract version from tag (usually format: v1.2.3 or 1.2.3)
	githubVersion := strings.TrimPrefix(release.TagName, "v")
	fmt.Printf("Latest version on GitHub : %s\n", githubVersion)

	// Compare versions
	if compareSemVer(githubVersion, currentVersion) <= 0 {
		fmt.Println("Service already using the latest version. No update needed.")
		return
	}

	// Download the newer version
	fmt.Printf("Newer version found! Downloading version %s...\n", githubVersion)
	_, err = downloadLatestJar(release)
	if err != nil {
		ExitWithError("Failed to download latest Komga jar", err)
	}

	// Find the latest jar (should be the one we just downloaded)
	latestJarPath, latestVersion, err := findLatestJar()
	if err != nil {
		ExitWithError("Failed to find latest Komga jar", err)
	}

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
}

// fetchLatestRelease fetches the latest release info from GitHub API
func fetchLatestRelease() (*GitHubRelease, error) {
	url := "https://api.github.com/repos/gotson/komga/releases/latest"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent header (required by GitHub API)
	req.Header.Set("User-Agent", "Komga-Updater")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadLatestJar downloads the jar file from the given release
func downloadLatestJar(release *GitHubRelease) (string, error) {
	// Find the jar asset
	var jarURL, jarName string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".jar") && strings.HasPrefix(asset.Name, "komga-") {
			jarURL = asset.BrowserDownloadURL
			jarName = asset.Name
			break
		}
	}

	if jarURL == "" {
		return "", fmt.Errorf("no jar file found in release %s", release.TagName)
	}

	// Check if jar already exists locally
	targetPath := filepath.Join(KomgaDir, jarName)
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf("Jar file %s already exists locally\n", jarName)
		return jarName, nil
	}

	// Download the jar
	fmt.Printf("Downloading %s...\n", jarName)
	if err := downloadFile(targetPath, jarURL); err != nil {
		return "", fmt.Errorf("failed to download jar: %w", err)
	}

	fmt.Printf("Successfully downloaded %s\n", jarName)
	return jarName, nil
}

// downloadFile downloads a file from url and saves it to filepath
func downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file with progress
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded %d bytes\n", written)
	return nil
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
