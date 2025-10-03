package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	Version          string
	KomgaServiceFile string
	KomgaDir         string
	ImmichDir        string
)

var (
	validServices = []string{"komga", "immich", "jellyfin"}
)

func ExitWithError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %s: %v\n", msg, err)
	os.Exit(1)
}

func printLogo() {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╭─────────────────────────────────────────────────────────────╮\n")
	fmt.Fprintf(os.Stderr, "│                                                             │\n")
	fmt.Fprintf(os.Stderr, "│              Media Server Service Updater %-17s │\n", "v"+Version)
	fmt.Fprintf(os.Stderr, "│                                                             │\n")
	fmt.Fprintf(os.Stderr, "╰─────────────────────────────────────────────────────────────╯\n")
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, "  A utility to automatically update media server services to\n")
	fmt.Fprintf(os.Stderr, "  their latest versions and clean up old artifacts.\n")
	fmt.Fprintf(os.Stderr, "\n")
}

func printConfig() {
	fmt.Fprintf(os.Stderr, "CONFIGURATION\n")
	fmt.Fprintf(os.Stderr, "  The following paths are used by this application:\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Komga:\n")
	fmt.Fprintf(os.Stderr, "    Service File: %s\n", KomgaServiceFile)
	fmt.Fprintf(os.Stderr, "    Directory:    %s\n", KomgaDir)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Immich:\n")
	fmt.Fprintf(os.Stderr, "    Directory:    %s\n", ImmichDir)
	fmt.Fprintf(os.Stderr, "\n")
}

func printUsage() {
	progName := filepath.Base(os.Args[0])

	printLogo()

	fmt.Fprintf(os.Stderr, "USAGE\n")
	fmt.Fprintf(os.Stderr, "  %s -service <name>\n", progName)
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, "OPTIONS\n")
	fmt.Fprintf(os.Stderr, "  -service <name>    Service to update (required)\n")
	fmt.Fprintf(os.Stderr, "  -h, -help          Show this help message\n")
	fmt.Fprintf(os.Stderr, "  -v, -version       Show version info\n")
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, "AVAILABLE SERVICES\n")
	for _, svc := range validServices {
		fmt.Fprintf(os.Stderr, "  • %s\n", svc)
	}
	fmt.Fprintf(os.Stderr, "\n")

	printConfig()

	fmt.Fprintf(os.Stderr, "EXAMPLES\n")
	fmt.Fprintf(os.Stderr, "  Update Komga to the latest version:\n")
	fmt.Fprintf(os.Stderr, "    $ %s -service komga\n", progName)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Update Immich to the latest version:\n")
	fmt.Fprintf(os.Stderr, "    $ %s -service immich\n", progName)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Show this help:\n")
	fmt.Fprintf(os.Stderr, "    $ %s -help\n", progName)
	fmt.Fprintf(os.Stderr, "\n")
}

func printVersion() {
	printLogo()
	printConfig()
}

func isValidService(service string) bool {
	if slices.Contains(validServices, service) {
		return true
	}

	return false
}

func main() {
	service := flag.String("service", "", fmt.Sprintf("Service to update (required): %s",
		strings.Join(validServices, ", ")))
	helpFlag := flag.Bool("help", false, "Show this help message")
	flag.BoolVar(helpFlag, "h", false, "Show this help message (shorthand)")
	versionFlag := flag.Bool("version", false, "Show version info")
	flag.BoolVar(versionFlag, "v", false, "Show version info (shorthand)")

	// Custom usage function
	flag.Usage = printUsage

	flag.Parse()

	// Show version if requested
	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	// Show help if requested
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Error: This application requires root privileges.\n")
		fmt.Fprintf(os.Stderr, "Please run with sudo:\n")
		fmt.Fprintf(os.Stderr, "  sudo %s -service <name>\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	// Validate service flag
	if *service == "" {
		fmt.Fprintf(os.Stderr, "Error: -service flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if !isValidService(*service) {
		fmt.Fprintf(os.Stderr, "Error: invalid service '%s'\n\n", *service)
		flag.Usage()
		os.Exit(1)
	}

	// Execute service update
	switch *service {
	case "komga":
		UpdateKomga()
	case "immich":
		UpdateImmich()
	case "jellyfin":
		UpdateJellyfin()
	}
}
