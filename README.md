# worklog

A fast CLI + TUI for logging work highlights and lowlights throughout the year, so nothing is forgotten at review time.

```
worklog highlight "Delivered the auth refactor two days ahead of schedule"
worklog lowlight "Missed the deploy window on Tuesday"
```

## Why

Performance reviews reward the people who remember what they did - not the people who did the most. `worklog` makes it frictionless to capture a win or a miss in seconds, and gives you a clean way to review them when it matters.

## Features

- **Fast capture** from the terminal - one command, no friction
- **Highlight / Lowlight** categorization out of the box
- **TUI** for review time: two-column view, navigate, edit, delete, filter, browse by quarter
- **Web UI** via `worklog serve` - same two-column view in the browser, auto-refreshes
- **Markdown export** for any quarter or full year, from CLI or TUI
- **SQLite storage** at `~/.worklog/worklog.db` - local, private, yours
- **Pre-built binaries** for Linux, macOS (Intel + Apple Silicon), and Windows

## Installation

### Homebrew (macOS and Linux)

```bash
brew install jquiaios/tap/worklog
```

### Go install

If you have Go 1.23+ installed:

```bash
go install github.com/jquiaios/worklog/cmd/worklog@latest
```

### Download a binary

Grab the latest release for your platform from the [Releases page](https://github.com/jquiaios/worklog/releases), extract it, and move it somewhere on your `$PATH`:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/jquiaios/worklog/releases/latest/download/worklog_Darwin_arm64.tar.gz | tar xz
sudo mv worklog /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/jquiaios/worklog/releases/latest/download/worklog_Darwin_amd64.tar.gz | tar xz
sudo mv worklog /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/jquiaios/worklog/releases/latest/download/worklog_Linux_amd64.tar.gz | tar xz
sudo mv worklog /usr/local/bin/
```

On macOS, if you downloaded the binary directly (not via Homebrew), Gatekeeper will block it on first run. Remove the quarantine flag once:

```bash
xattr -d com.apple.quarantine /usr/local/bin/worklog
```

## Usage

### Add an entry

All of the following are equivalent:

```bash
worklog highlight "Won the sprint demo with the client"
worklog hl "Won the sprint demo with the client"
worklog add highlight "Won the sprint demo with the client"
worklog add hl "Won the sprint demo with the client"
```

```bash
worklog lowlight "Broke prod on Friday afternoon"
worklog ll "Broke prod on Friday afternoon"
worklog add lowlight "Broke prod on Friday afternoon"
worklog add ll "Broke prod on Friday afternoon"
```

### List entries

```bash
worklog list                    # all entries, newest first
worklog list -t hl              # highlights only
worklog list --type lowlight    # lowlights only
```

Output:
```
#4    2026-04-18  [highlight]  Won the sprint demo with the client
#3    2026-04-17  [lowlight]   Broke prod on Friday afternoon
```

### Delete an entry

```bash
worklog delete 3
worklog rm 3      # shorthand
```

### Export to Markdown

```bash
worklog export                        # current quarter, printed to stdout
worklog export -o review.md           # write to file instead

worklog export -q Q1                  # specific quarter (current year)
worklog export -q Q4 -y 2025          # specific quarter of a past year
worklog export -y 2025                # full year (entries grouped by quarter)
worklog export -y 2025 -o 2025.md     # full year, written to file
```

Output is a Markdown file with Highlights and Lowlights sections. Full-year exports include `### Q1 2025` subheadings inside each section.

### Start the web UI

```bash
worklog serve             # opens http://localhost:7171 in your browser
worklog serve -p 8080     # use a different port if 7171 is taken
worklog serve --no-open   # start the server without opening the browser
```

The web UI auto-refreshes every 3 seconds, so entries added from the CLI or TUI appear automatically.

### Open the TUI

```bash
worklog
```

Launches a full-terminal two-column view of your entries, scoped to the current quarter by default.

**Navigation**

| Key | Action |
|-----|--------|
| `tab` | Switch between Highlights and Lowlights |
| `[` | Previous quarter |
| `]` | Next quarter |
| `a` | Show all entries (no period filter) |

**Actions** 

| Key | Action |
|-----|--------|
| `n` | New entry in the focused column |
| `e` | Edit selected entry |
| `d` | Delete selected entry (asks for confirmation) |
| `x` | Export current period to Markdown file |
| `/` | Filter entries in the focused column |
| `q` | Quit |

## Data

Entries are stored in a SQLite database at `~/.worklog/worklog.db`. It is never sent anywhere. You can back it up, copy it to another machine, or open it directly with any SQLite client.

## Development

The project uses Docker so you don't need Go installed locally.

```bash
git clone https://github.com/jquiaios/worklog.git
cd worklog
docker build -t worklog .
mkdir -p ~/.worklog
docker run -it --rm --user $(id -u):$(id -g) -v ~/.worklog:/home/wl/.worklog worklog

# For the web UI inside Docker, publish the port and bind to all interfaces
# inside the container (Docker's port forwarding is the isolation boundary):
docker run --rm -p 7171:7171 -v ~/.worklog:/home/wl/.worklog worklog serve --host 0.0.0.0 --no-open
```

After adding or removing dependencies, regenerate `go.sum`:

```bash
docker run --rm -v $(pwd):/src -w /src golang:1.23-alpine go mod tidy
```

## Releasing

Releases are automated via GoReleaser and GitHub Actions. To cut a new release:

```bash
git tag v1.2.3
git push origin v1.2.3
```

The workflow builds binaries for all platforms and publishes a GitHub Release automatically.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - TUI styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [modernc/sqlite](https://gitlab.com/cznic/sqlite) - Pure-Go SQLite driver

## License

MIT
