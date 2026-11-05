# Caido Plugin Installer

A small Go tool that installs Caido plugins from a `.zip` archive file and moves the built files into the Caido plugins directory.
During the installation process it will also append the necessary rows to the Caido SQLite database.

## What it does

1. Takes a plugin archive (`.zip` file) as input
2. Extracts the archive to a temp folder
3. Reads the `manifest.json` to get plugin info (name, version, author, etc.)
4. Generates UUIDs for the backend, frontend, and package ID
5. Copies backend and frontend files to `~/.local/share/caido/plugins/`
6. Stores plugin info in a local SQLite database (`~/.local/share/caido/plugins.db`)

## Usage

```bash
go run main.go -package plugin.zip
```

Or build it first:

```bash
go build -o caido_plugin_installer .
./caido_plugin_installer -package plugin.zip
```

## How it works (quick overview)

| File | What it does |
|------|-------------|
| `main.go` | Entry point — parses flags, orchestrates the install |
| `cmd/archive.go` | Extracts the `.zip` file |
| `cmd/move.go` | Copies plugin files to the right folders |
| `utils/manifest.go` | Parses the `manifest.json` into a Go struct |
| `utils/sqlite.go` | Inserts plugin data into the Caido SQLite DB |
| `utils/uuid.go` | Generates UUIDs for plugin IDs |

## Requirements

- Go 1.27+
- SQLite3 (comes with the `go-sqlite3` dependency)
- `pnpm` to build the plugin.  Can be installed with `npm install -g pnpm`
