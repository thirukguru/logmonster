# Log Monster Detector 🦖

A lightweight CLI utility for Linux/MacOS that detects processes and services writing excessively to log files and consuming disk space.

## Features

- **Scan** - Detect rapidly growing log files with snapshot comparison
- **Watch** - Real-time monitoring with live-updating TUI
- **Blame** - Map log files to the processes writing to them
- **Service** - Resolve PIDs to systemd service units
- **Top** - Rank processes by disk write rate
- **Kill** - Safely terminate runaway processes

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/thirukguru/logmonster/main/install.sh | bash
```

### From Source

```bash
git clone https://github.com/thirukguru/logmonster.git
cd logmonster
make build
sudo mv bin/logmonster /usr/local/bin/
```

### Pre-built Binaries

Download from the [Releases](https://github.com/thiruk/logmonster/releases) page.

## Quick Start

```bash
# Scan /var/log for rapidly growing files
logmonster scan

# Watch files in real-time
logmonster watch

# Find which process is writing to a log file
logmonster blame /var/log/apache2/error.log

# Find the systemd service for a PID
logmonster service 4321

# Show top I/O offenders
logmonster top

# Kill a runaway process (requires sudo)
sudo logmonster kill 4321
```

## Commands

### `logmonster scan`

Take two snapshots of file sizes and calculate growth rates.

```bash
logmonster scan [options]

Options:
  --paths      Comma-separated paths to scan (default: /var/log,/tmp)
  --interval   Seconds between snapshots (default: 5)
  --threshold  Minimum growth in MB to report (default: 10)
```

### `logmonster watch`

Continuously monitor file growth in real-time with a TUI.

```bash
logmonster watch [options]

Options:
  --refresh   Update interval in seconds (default: 2)
  --top       Number of files to display (default: 10)
```

### `logmonster blame <file>`

Identify which process is writing to a specific file.

```bash
logmonster blame /var/log/myapp.log
```

### `logmonster service <pid>`

Resolve a PID to its systemd service unit.

```bash
logmonster service 4321
```

### `logmonster top`

Show top processes by write rate.

```bash
logmonster top [options]

Options:
  --limit     Number of processes to show (default: 10)
  --interval  Measurement period in seconds (default: 5)
```

### `logmonster kill <pid>`

Terminate a process with graceful shutdown.

```bash
sudo logmonster kill 4321 [options]

Options:
  --force     Skip confirmation prompt
  --timeout   Seconds to wait before SIGKILL (default: 5)
```

## Configuration

Create `~/.logmonster/config.yaml`:

```yaml
scan_paths:
  - /var/log
  - /tmp

exclude_patterns:
  - "*.gz"
  - "*.zip"

thresholds:
  growth_mb: 10
  rate_mb_per_sec: 1.0

display:
  top_n: 10
  use_colors: true
```

## Exit Codes

| Code | Meaning           |
|------|-------------------|
| 0    | Success           |
| 1    | Invalid arguments |
| 2    | Permission error  |
| 3    | Target not found  |
| 4    | Config error      |
| 130  | User interrupt    |

## Requirements

- Linux (uses `/proc` filesystem and systemd)
- Go 1.21+ (for building from source)
- `lsof` command (for blame functionality)

## License

MIT License
