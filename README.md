# klip – a tiny cross‑platform clipboard manager

`klip` lets you store short text snippets under a name (key) and instantly copy them to the system clipboard.  
It is a single binary with no background daemon, stores everything in a JSON file, and works on macOS,
Linux and Windows.

---

## Installation

### From Source (Recommended)

```bash
git clone https://github.com/suprbdev/klip.git
cd klip
make install   # or: go install ./cmd/klip
```

### Using Go

```bash
go install github.com/suprbdev/klip/cmd/klip@latest   # requires Go 1.22+
```

The binary will be placed in `$GOPATH/bin` (or `$HOME/go/bin`).  
Add that directory to your `PATH` or copy the binary wherever you like.

**Quick start**: Run `klip help` to see all available commands and `klip help <command>` for detailed help on any command.

---

## Usage

```bash
# Save a value and automatically copy it
klip set api-token abc123

# Save with an expiration time (10 minutes) - NOTE: flags must come before positional args
klip set --ttl 10m note "hello world"

# Read from stdin, useful in pipelines
echo secret | klip put tmp --temp   # default temp TTL = 24h (configurable)

# Retrieve a value (copies to clipboard by default)
klip get api-token

# Same as `get` but without copying (recommended to avoid clipboard issues)
klip get api-token --no-clip

# Print value without copying to clipboard
klip print api-token

# Copy to clipboard without printing the value
klip get api-token --silent

# List stored keys
klip ls               # only active entries
klip ls --all         # include expired, marked with "(expired)"
klip ls --long        # show timestamps

# Remove one or more keys
klip rm foo bar

# Delete all expired entries
klip clear --expired

# Wipe the whole store
klip clear --all

# Alias for `get` that always copies
klip cp api-token

# Copy to clipboard silently (no output)
klip cp api-token --silent

# Print value without copying (useful for scripts)
klip print api-token
klip print api-token --raw

# Rename a key (use --force to overwrite)
klip mv old-key new-key
klip mv secret password --force

# Export / import the whole JSON store
klip export > backup.json
cat backup.json | klip import

# Show where the store lives on disk
klip config --path

# Quick statistics
klip stats

# Get help
klip help                    # Show main help
klip help set               # Show help for 'set' command
klip -h                     # Alternative help flag
klip --help                 # Alternative help flag
```

### Getting Help

klip provides comprehensive help documentation:

```bash
# Show main help with all available commands
klip help
klip -h
klip --help

# Show detailed help for a specific command
klip help set
klip help get
klip help ls
# ... and so on for any command
```

### Flags common to many commands

| Flag          | Meaning |
|---------------|---------|
| `--no-clip`   | Do **not** copy to clipboard (default for `get` is to copy). |
| `--raw`       | Print value without trailing newline (get/print/cp commands only). |
| `--silent`    | Copy to clipboard without printing the value (get/cp commands only). |
| `--ttl <dur>` | Set expiration, e.g. `10m`, `2h`, `1d`. |
| `--temp`      | Shortcut for a temporary entry; TTL defaults to 24 hours (configurable). |
| `--from-stdin`| Read the value from STDIN instead of as an argument. |

### Environment variables

| Variable                     | Effect |
|------------------------------|--------|
| `KLIP_DISABLE_CLIPBOARD=1`   | Globally disable clipboard copying (flags still override). |
| `KLIP_DEFAULT_TEMP_TTL=2h`   | Change the default TTL used with `--temp`. |
| `KLIP_CONFIG_DIR=/tmp/klip`  | Override where the JSON store is kept (useful for testing). |

---

## Design notes

* **Storage** – a single JSON file (`$XDG_CONFIG_HOME/klip/store.json` on Linux, `%APPDATA%/klip/store.json` on Windows, `~/Library/Application Support/klip/store.json` on macOS).  
  The file is written atomically using a temporary file + rename.

* **Expiration** – an entry with `ExpiresAt` in the past is considered expired and ignored by default.  

* **Clipboard** – powered by the well‑tested `github.com/atotto/clipboard`. If copying fails, the value is printed to STDOUT and a warning goes to STDERR.

* **Concurrency** – the store file is small; we rely on atomic writes and the OS’s rename semantics for safety.  
  For very high parallelism a lockfile could be added later.

---

## Building from source

### Using Make (Recommended)

```bash
git clone https://github.com/suprbdev/klip.git
cd klip

# Build for current platform
make build

# Build for all platforms
make build-all

# Install to GOPATH/bin
make install

# Run tests
make test

# See all available targets
make help
```

### Using Go directly

```bash
git clone https://github.com/suprbdev/klip.git
cd klip
go build ./cmd/klip   # produces ./klip
```

The binary works on all supported platforms without external dependencies.

### Available Make targets

- `make build` - Build for current platform
- `make build-all` - Build for Linux, macOS, and Windows
- `make install` - Install to GOPATH/bin
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make test-binary` - Test the built binary functionality
- `make lint` - Run linter
- `make fmt` - Format code
- `make clean` - Remove build artifacts
- `make help` - Show all available targets

---

## Troubleshooting

### Clipboard Issues

If commands hang when trying to access the clipboard, use the `--no-clip` flag:

```bash
# Instead of: klip get mykey
klip get mykey --no-clip

# Instead of: klip set mykey "value"
klip set mykey "value" --no-clip
```

### Command Line Argument Order

For TTL functionality to work correctly, flags must come before positional arguments:

```bash
# ✅ Correct
klip set --ttl 10m --no-clip mykey "value"

# ❌ Incorrect (TTL won't work)
klip set mykey "value" --ttl 10m --no-clip
```

### Environment Variables

You can disable clipboard functionality globally:

```bash
export KLIP_DISABLE_CLIPBOARD=1
# Now all commands will skip clipboard operations
```

### Command Differences

Understanding when to use each command:

- **`klip get`** - Retrieves value and copies to clipboard (default behavior)
- **`klip print`** - Prints value without copying to clipboard (useful for scripts)
- **`klip cp`** - Alias for `get` (always copies to clipboard)

### Getting More Help

If you need more detailed information about any command:

```bash
# Show help for any command
klip help <command>

# Examples
klip help set
klip help get
klip help print
klip help ls
```

---

## License

MIT – see LICENSE file.

--- 

Enjoy using **klip**! 🎉
