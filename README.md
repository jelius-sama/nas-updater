# Media Server Service Updater

A command-line utility to automatically update media server services (Komga, Immich, Jellyfin) to their latest versions and clean up old artifacts.

## ⚠️ Important Notice

**This CLI tool is specifically designed for my personal server configuration and workflow.** It will only work for others if your setup matches mine exactly. This is not a general-purpose tool and assumes very specific directory structures, installation methods, and service configurations.

## System Requirements

- **Operating System**: Linux-based OS with systemd (tested on Ubuntu/Debian-based distributions, but should work on any Linux distribution with systemd)
- **Privileges**: Root/sudo access required for service management and file operations
- **Go**: Version 1.24.5 or higher (for building from source)

## My Configuration

This tool is built around the following specific setup:

### Jellyfin
- **Installation Method**: APT package manager (`apt install jellyfin`)
- **Service Management**: systemd service controlled via `apt upgrade jellyfin`
- **Notes**: Can be adapted to work with other package managers (yum, dnf, pacman, etc.) with minor modifications

### Komga
- **Installation Method**: Manual JAR file deployment
- **Directory**: `/opt/komga` (configurable via build flags)
- **Service File**: `/etc/systemd/system/komga.service` (configurable via build flags)
- **Update Process**: 
  - Automatically downloads the latest JAR from GitHub releases
  - Updates the systemd service file to point to the new version
  - Reloads and restarts the service
  - Cleans up old JAR files to save disk space
- **Version Detection**: Parses semver from JAR filenames (`komga-X.Y.Z.jar`)

### Immich
- **Installation Method**: Docker container
- **Directory**: `/opt/immich` (configurable via build flags)
- **Update Process**:
  - Stops the running Docker container
  - Pulls the latest image from Docker Hub
  - Starts the container with `docker-compose up -d`
  - Removes stale/unused Docker images to free up storage

## Installation

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/jelius-sama/nas-updater
cd nas-updater
```

2. Build with default configuration:
```bash
make host_default
```

3. Build with custom configuration (using `config.json`):
```bash
# Edit config.json with your paths
make host_default
```

The Makefile uses `ldflags` to inject configuration values from `config.json` at build time, allowing you to customize paths without modifying the source code.

### Configuration File Format

Modify the `config.json` file in the project root:

```json
{
    "Version": "1.0.0",
    "KomgaServiceFile": "/etc/systemd/system/komga.service",
    "KomgaDir": "/opt/komga",
    "ImmichDir": "/opt/immich"
}

```

## Usage

### Display Help

```bash
./nas-updater -help
```

or

```bash
./nas-updater -h
```

### Update a Service

**Note**: Root privileges are required for updating services (except when displaying help).

```bash
sudo ./nas-updater -service komga
```

```bash
sudo ./nas -service immich
```

```bash
sudo ./nas -service jellyfin
```

### Example Output

```
╭─────────────────────────────────────────────────────────────╮
│                                                             │
│              Media Server Service Updater v1.0.0            │
│                                                             │
╰─────────────────────────────────────────────────────────────╯

Current version in service: 1.10.0
Latest version available : 1.11.2
Updating service file to use version 1.11.2
Service updated and restarted with version 1.11.2
Deleting old jar: /opt/komga/komga-1.10.0.jar
Deleted 1 stale jar file(s)
```

## How It Works

### Komga Update Process

1. Scans `/opt/komga` for all JAR files matching `komga-*.jar`
2. Parses version numbers using semantic versioning
3. Identifies the latest version available
4. Reads the current version from the systemd service file
5. Compares versions and updates if a newer version is found
6. Updates the `ExecStart` line in the service file
7. Executes `systemctl daemon-reload` and `systemctl restart komga`
8. Deletes all old JAR files, keeping only the latest

### Immich Update Process

1. Navigates to the Immich Docker directory (`/opt/immich`)
2. Stops the running container using `docker-compose down`
3. Pulls the latest images with `docker-compose pull`
4. Starts the updated containers with `docker-compose up -d`
5. Removes unused/dangling Docker images to save space

### Jellyfin Update Process

1. Executes `apt update` to refresh package lists
2. Runs `apt upgrade jellyfin -y` to install the latest version
3. Restarts the Jellyfin service if needed

## Compatibility

This tool should work on **any Linux distribution with systemd**, including but not limited to:

- Ubuntu / Debian
- Fedora / RHEL / CentOS
- Arch Linux
- openSUSE
- Gentoo (with systemd)

**The key requirement is that your configuration matches mine:**
- Same directory structures
- systemd for service management
- Docker/Docker Compose for Immich (if using)
- Similar service file formats for Komga

## Customization

If your configuration differs from mine, you'll need to:

1. Modify the `config.json` with your paths
2. Rebuild using `make build`
3. Potentially adjust the update logic in the source code for different installation methods

## Limitations

- **Not plug-and-play**: Requires exact configuration matching
- **Jellyfin**: Currently hardcoded for APT; other package managers need code changes
- **No automatic configuration detection**: Paths must be set at build time
- **Linux-only**: Requires systemd and standard Linux utilities
- **No rollback mechanism**: Failed updates must be fixed manually

## License

This is a personal utility tool. Use at your own risk.

## Contributing

As this is a personal tool tailored to my specific setup, I'm not actively seeking contributions. However, feel free to fork and adapt it to your own needs!

## Troubleshooting

### "This application requires root privileges"

Run the command with `sudo`:
```bash
sudo ./nas-updater -service komga
```

### Service fails to restart

Check the service logs:
```bash
sudo journalctl -u komga -n 50
```

### Docker commands fail (Immich)

Ensure Docker and Docker Compose are installed and you have the correct permissions:
```bash
docker --version
docker-compose --version
```

### Wrong paths detected

Rebuild with your custom `config.json`:
```bash
make clean
make host_default
```

## Version

Current version: **1.0.0**

---

**Made with ☕ for personal server management**
