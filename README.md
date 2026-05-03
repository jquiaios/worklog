# worklog

[![CI](https://github.com/jquiaios/worklog/actions/workflows/ci.yml/badge.svg)](https://github.com/jquiaios/worklog/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/release/jquiaios/worklog.svg)](https://github.com/jquiaios/worklog/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/jquiaios/worklog.svg)](https://pkg.go.dev/github.com/jquiaios/worklog)
[![Go Report Card](https://goreportcard.com/badge/github.com/jquiaios/worklog)](https://goreportcard.com/report/github.com/jquiaios/worklog)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Log work highlights and lowlights as they happen — so performance reviews aren't a memory test.

## Why not just use Obsidian, Notion, or a text file?

**Obsidian** and **Notion** are great for notes you *plan* to write.
The problem is that by the time you've switched context, found the right file, and figured out where to put the thing, the moment has passed — and it never gets written.

**A text file** is closer, but you still have to find it, open it, and maintain some structure yourself.

**worklog** lives in your terminal: one command, no context switch, no decisions.
That gap between *"I'll note that later"* and *"I actually note it"* is exactly what it closes.

![worklog TUI demo](./docs/demo.gif)

<!-- TODO: replace docs/demo.gif with a VHS-generated GIF of the TUI: open → add a highlight → add a lowlight → navigate a quarter → quit -->

> The example above shows the TUI. See [`docs/demo.tape`](./docs/demo.tape) for the source.

## Tutorial

To get started, [install worklog](#installation).

Capture a highlight the moment it happens:

```bash
worklog hl "Delivered the auth refactor two days ahead of schedule"
```

Capture a lowlight just as quickly:

```bash
worklog ll "Missed the deploy window on Tuesday"
```

When review season comes around, open the TUI to browse your quarter:

```bash
worklog
```

Or export to Markdown for your review doc:

```bash
# current quarter
worklog export -o review-q2-2026.md

# Or a full year:
worklog export -y 2025 -o review-2025.md
```

That's it. Keep logging as you go, and you'll never walk into a 1:1 empty-handed again.

## Installation

> **Note**
> worklog stores its data in a SQLite database at `~/.worklog/worklog.db`. Nothing is ever sent over the network.

Use a package manager:

```bash
# macOS or Linux
brew install jquiaios/tap/worklog
```

Or install with `go` (requires Go 1.25+):

```bash
go install github.com/jquiaios/worklog/cmd/worklog@latest
```

Or use the install script (detects your OS and architecture automatically):

```bash
curl -fsSL https://raw.githubusercontent.com/jquiaios/worklog/main/install.sh | sh
```

To pin a specific version or choose a different install directory:

```bash
WORKLOG_VERSION=0.2.0 curl -fsSL https://raw.githubusercontent.com/jquiaios/worklog/main/install.sh | sh
WORKLOG_INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/jquiaios/worklog/main/install.sh | sh
```

Or download a pre-built binary directly from the [Releases page](https://github.com/jquiaios/worklog/releases) for Linux, macOS (Intel + Apple Silicon), or Windows.

> **macOS note:** If you installed via the script or downloaded directly (not Homebrew), Gatekeeper may block the first run. Remove the quarantine flag once:
> ```bash
> xattr -d com.apple.quarantine /usr/local/bin/worklog
> ```

## CLI

The CLI is designed to be fast enough that logging an entry never breaks your flow.

### Add an entry

All of the following are equivalent:

```bash
worklog highlight "Won the sprint demo with the client"
# or
worklog hl "Won the sprint demo with the client"
# or
worklog add highlight "Won the sprint demo with the client"
# or
worklog add hl "Won the sprint demo with the client"
```

The same shorthands apply to lowlights (`lowlight`, `ll`, `add lowlight`, `add ll`).

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

Exports are plain Markdown with `Highlights` and `Lowlights` sections. Full-year exports add `### Q1 2025` subheadings inside each section.

## TUI

Run `worklog` with no arguments to launch the TUI — a two-column view scoped to the current quarter by default.

![worklog TUI](./docs/tui.gif)

> See [`docs/tui.tape`](./docs/tui.tape) for the source.

**Navigation**

| Key   | Action                                |
| ----- | ------------------------------------- |
| `tab` | Switch between Highlights and Lowlights |
| `[`   | Previous quarter                      |
| `]`   | Next quarter                          |
| `a`   | Show all entries (no period filter)   |

**Actions**

| Key | Action                                          |
| --- | ----------------------------------------------- |
| `n` | New entry in the focused column                 |
| `e` | Edit selected entry                             |
| `d` | Delete selected entry (asks for confirmation)   |
| `x` | Export current period to Markdown file          |
| `/` | Filter entries in the focused column            |
| `q` | Quit                                            |

## Web UI

Prefer a browser? `worklog serve` starts a local web UI that mirrors the TUI's two-column layout and auto-refreshes every 3 seconds — so entries added from the CLI or TUI appear without reloading.

```bash
worklog serve             # opens http://localhost:7171 in your browser
worklog serve -p 8080     # use a different port if 7171 is taken
worklog serve --no-open   # start without opening the browser
```

![worklog web UI](./docs/web.png)

<!-- TODO: replace docs/web.png with a screenshot of the web UI -->

## Data

Entries live in a SQLite database at `~/.worklog/worklog.db`. It's local, private, and yours — you can back it up, copy it to another machine, or open it with any SQLite client.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — TUI styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [modernc/sqlite](https://gitlab.com/cznic/sqlite) — Pure-Go SQLite driver

## Contributing

Issues and pull requests are welcome. See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and guidelines.

## License

[MIT](./LICENSE)