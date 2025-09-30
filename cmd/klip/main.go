package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	atottoclipboard "github.com/atotto/clipboard"
	"github.com/suprbdev/klip/internal/store"
	"github.com/suprbdev/klip/internal/timeutil"
)

const (
	exitOK          = 0
	exitError       = 1
	exitNotFound    = 2
	envDisableClip  = "KLIP_DISABLE_CLIPBOARD"
	envDefaultTTL   = "KLIP_DEFAULT_TEMP_TTL"
	envConfigDir    = "KLIP_CONFIG_DIR" // for tests / custom location
	clipboardFailed = "WARNING: clipboard unavailable – printed to STDOUT\n"
)

func main() {
	if len(os.Args) < 2 {
		showMainHelp()
		os.Exit(exitOK)
	}
	cmd := os.Args[1]
	switch cmd {
	case "help", "-h", "--help":
		if len(os.Args) > 2 {
			showSubcommandHelp(os.Args[2])
		} else {
			showMainHelp()
		}
	case "set":
		handleSet(os.Args[2:])
	case "get":
		handleGet(os.Args[2:])
	case "rm":
		handleRm(os.Args[2:])
	case "ls":
		handleLs(os.Args[2:])
	case "clear":
		handleClear(os.Args[2:])
	case "cp":
		handleCp(os.Args[2:])
	case "put":
		handlePut(os.Args[2:])
	case "mv":
		handleMv(os.Args[2:])
	case "export":
		handleExport()
	case "import":
		handleImport()
	case "config":
		handleConfig(os.Args[2:])
	case "stats":
		handleStats()
	case "print":
		handlePrint(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", cmd)
		fmt.Fprintf(os.Stderr, "Run 'klip help' for usage information.\n")
		os.Exit(exitError)
	}
}

/*** Helpers ***/

func getStorePath() string {
	if p := os.Getenv(envConfigDir); p != "" {
		return filepath.Join(p, "store.json")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot determine config dir:", err)
		os.Exit(exitError)
	}
	return filepath.Join(dir, "klip", "store.json")
}

func loadStore() *store.Store {
	path := getStorePath()
	s, err := store.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load store:", err)
		os.Exit(exitError)
	}
	return s
}

func saveStore(s *store.Store) {
	path := getStorePath()
	if err := store.Save(path, s); err != nil {
		fmt.Fprintln(os.Stderr, "failed to save store:", err)
		os.Exit(exitError)
	}
}

// copyToClipboard copies text if clipboard is enabled.
// Returns true when it succeeded, false otherwise.
func copyToClipboard(text string) bool {
	if atottoclipboard.Unsupported {
		return false
	}
	if err := atottoclipboard.WriteAll(text); err != nil {
		fmt.Fprint(os.Stderr, clipboardFailed)
		return false
	}
	return true
}

/*** Help Functions ***/

func showMainHelp() {
	fmt.Printf(`klip - A tiny cross-platform clipboard manager

USAGE:
    klip <COMMAND>

COMMANDS:
    set      Store a value under a key
    get      Retrieve and copy a value to clipboard
    print    Print a value without copying to clipboard
    rm       Remove one or more keys
    ls       List stored keys
    clear    Remove expired entries or wipe the store
    cp       Alias for 'get' (always copies to clipboard)
    put      Store value from stdin
    mv       Rename a key
    export   Export the store as JSON
    import   Import store from JSON on stdin
    config   Show configuration information
    stats    Show store statistics
    help     Show this help or help for a subcommand

FLAGS:
    -h, --help    Show help information

EXAMPLES:
    klip set api-token abc123
    klip get api-token --no-clip
    klip print api-token
    klip set --ttl 10m --no-clip note "hello world"
    klip ls --long
    klip help set

ENVIRONMENT VARIABLES:
    KLIP_DISABLE_CLIPBOARD=1     Disable clipboard operations globally
    KLIP_DEFAULT_TEMP_TTL=2h     Default TTL for --temp flag
    KLIP_CONFIG_DIR=/path        Override store location

For more information about a specific command, run 'klip help <COMMAND>'.
`)
}

func showSubcommandHelp(subcommand string) {
	switch subcommand {
	case "set":
		fmt.Printf(`klip set - Store a value under a key

USAGE:
    klip set [FLAGS] <KEY> [VALUE]

ARGUMENTS:
    <KEY>     The key to store the value under
    <VALUE>   The value to store (optional if --from-stdin is used)

FLAGS:
    --ttl <DURATION>     Set expiration time (e.g., 10m, 2h, 1d)
    --temp               Use default temporary TTL (24h by default)
    --no-clip            Do not copy to clipboard
    --from-stdin         Read value from stdin instead of argument

EXAMPLES:
    klip set api-token abc123
    klip set --ttl 10m --no-clip note "hello world"
    klip set --temp --no-clip temp-key "temporary data"
    echo "secret" | klip set --from-stdin secret-key
`)
	case "get":
		fmt.Printf(`klip get - Retrieve and copy a value to clipboard

USAGE:
    klip get [FLAGS] <KEY>

ARGUMENTS:
    <KEY>     The key to retrieve

FLAGS:
    --no-clip    Do not copy to clipboard
    --raw        Print value without trailing newline
    --silent     Copy to clipboard without printing value

EXAMPLES:
    klip get api-token
    klip get api-token --no-clip
    klip get api-token --raw
    klip get api-token --silent
`)
	case "print":
		fmt.Printf(`klip print - Print a value without copying to clipboard

USAGE:
    klip print [FLAGS] <KEY>

ARGUMENTS:
    <KEY>     The key to retrieve

FLAGS:
    --raw        Print value without trailing newline

EXAMPLES:
    klip print api-token
    klip print api-token --raw
`)
	case "rm":
		fmt.Printf(`klip rm - Remove one or more keys

USAGE:
    klip rm <KEY>...

ARGUMENTS:
    <KEY>...    One or more keys to remove

EXAMPLES:
    klip rm old-key
    klip rm key1 key2 key3
`)
	case "ls":
		fmt.Printf(`klip ls - List stored keys

USAGE:
    klip ls [FLAGS]

FLAGS:
    --all     Include expired entries (marked with "(expired)")
    --long    Show timestamps and expiration times

EXAMPLES:
    klip ls
    klip ls --all
    klip ls --long
`)
	case "clear":
		fmt.Printf(`klip clear - Remove expired entries or wipe the store

USAGE:
    klip clear [FLAGS]

FLAGS:
    --expired    Remove only expired entries
    --all        Wipe the entire store

EXAMPLES:
    klip clear --expired
    klip clear --all
`)
	case "cp":
		fmt.Printf(`klip cp - Alias for 'get' (always copies to clipboard)

USAGE:
    klip cp [FLAGS] <KEY>

ARGUMENTS:
    <KEY>     The key to retrieve

FLAGS:
    --no-clip    Do not copy to clipboard
    --raw        Print value without trailing newline
    --silent     Copy to clipboard without printing value

EXAMPLES:
    klip cp api-token
    klip cp api-token --raw
    klip cp api-token --silent
`)
	case "put":
		fmt.Printf(`klip put - Store value from stdin

USAGE:
    klip put [FLAGS] <KEY>

ARGUMENTS:
    <KEY>     The key to store the value under

FLAGS:
    --ttl <DURATION>     Set expiration time (e.g., 10m, 2h, 1d)
    --temp               Use default temporary TTL (24h by default)

EXAMPLES:
    echo "secret" | klip put secret-key
    echo "data" | klip put --ttl 1h temp-data
    cat file.txt | klip put --temp file-content
`)
	case "mv":
		fmt.Printf(`klip mv - Rename a key

USAGE:
    klip mv [FLAGS] <OLD_KEY> <NEW_KEY>

ARGUMENTS:
    <OLD_KEY>    The existing key to rename
    <NEW_KEY>    The new key name

FLAGS:
    --force    Overwrite destination if it exists

EXAMPLES:
    klip mv old-key new-key
    klip mv secret password --force
`)
	case "export":
		fmt.Printf(`klip export - Export the store as JSON

USAGE:
    klip export

DESCRIPTION:
    Exports the entire store as JSON to stdout. This can be used for
    backup purposes or to migrate data.

EXAMPLES:
    klip export > backup.json
    klip export | gzip > backup.json.gz
`)
	case "import":
		fmt.Printf(`klip import - Import store from JSON on stdin

USAGE:
    klip import

DESCRIPTION:
    Imports entries from JSON on stdin, merging them with the existing store.
    Entries with the same key will be overwritten.

EXAMPLES:
    cat backup.json | klip import
    gunzip -c backup.json.gz | klip import
`)
	case "config":
		fmt.Printf(`klip config - Show configuration information

USAGE:
    klip config [FLAGS]

FLAGS:
    --path    Show the store file path

EXAMPLES:
    klip config --path
`)
	case "stats":
		fmt.Printf(`klip stats - Show store statistics

USAGE:
    klip stats

DESCRIPTION:
    Shows the number of active and expired entries in the store.

EXAMPLES:
    klip stats
`)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		fmt.Fprintf(os.Stderr, "Run 'klip help' to see available commands.\n")
		os.Exit(exitError)
	}
}

/*** Subcommands ***/

func handleSet(args []string) {
	fs := flag.NewFlagSet("set", flag.ExitOnError)
	var ttlStr string
	var temp bool
	var noClip bool
	var fromStdin bool
	fs.StringVar(&ttlStr, "ttl", "", "time‑to‑live (e.g. 10m, 2h, 1d)")
	fs.BoolVar(&temp, "temp", false, "shortcut for --ttl $KLIP_DEFAULT_TEMP_TTL (default 24h)")
	fs.BoolVar(&noClip, "no-clip", false, "do not copy to clipboard")
	fs.BoolVar(&fromStdin, "from-stdin", false, "read value from STDIN")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "set requires a key")
		os.Exit(exitError)
	}
	key := rest[0]

	var value string
	if fromStdin || (len(rest) == 1 && !fromStdin) {
		// read whole stdin
		b, err := io.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to read stdin:", err)
			os.Exit(exitError)
		}
		value = strings.TrimRight(string(b), "\n")
	} else if len(rest) >= 2 {
		value = rest[1]
	} else {
		fmt.Fprintln(os.Stderr, "value missing; use --from-stdin or supply as argument")
		os.Exit(exitError)
	}

	var expiresAt *time.Time
	if ttlStr != "" && temp {
		fmt.Fprintln(os.Stderr, "--ttl and --temp are mutually exclusive")
		os.Exit(exitError)
	}
	if temp {
		def := os.Getenv(envDefaultTTL)
		if def == "" {
			def = "24h"
		}
		dur, err := timeutil.ParseTTL(def)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid default ttl:", err)
			os.Exit(exitError)
		}
		t := time.Now().Add(dur)
		expiresAt = &t
	} else if ttlStr != "" {
		dur, err := timeutil.ParseTTL(ttlStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid ttl:", err)
			os.Exit(exitError)
		}
		t := time.Now().Add(dur)
		expiresAt = &t
	}

	entry := store.Entry{
		Value:     value,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Temp:      temp,
	}
	s := loadStore()
	s.Set(key, entry)
	saveStore(s)

	if !noClip {
		copyToClipboard(value)
	}
}

func handleGet(args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	var noClip bool
	var raw bool
	var silent bool
	fs.BoolVar(&noClip, "no-clip", false, "do not copy to clipboard")
	fs.BoolVar(&raw, "raw", false, "print value without newline")
	fs.BoolVar(&silent, "silent", false, "copy to clipboard without printing value")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "get requires a key")
		os.Exit(exitError)
	}
	key := rest[0]

	s := loadStore()
	entry, ok := s.Get(key)
	if !ok {
		fmt.Fprintf(os.Stderr, "key %q not found or expired\n", key)
		os.Exit(exitNotFound)
	}

	if !noClip {
		copyToClipboard(entry.Value)
	}

	// Only print if not silent
	if !silent {
		if raw {
			fmt.Print(entry.Value)
		} else {
			fmt.Println(entry.Value)
		}
	}
}

func handlePrint(args []string) {
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	var raw bool
	fs.BoolVar(&raw, "raw", false, "print value without newline")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "print requires a key")
		os.Exit(exitError)
	}
	key := rest[0]

	s := loadStore()
	entry, ok := s.Get(key)
	if !ok {
		fmt.Fprintf(os.Stderr, "key %q not found or expired\n", key)
		os.Exit(exitNotFound)
	}

	// Print the value (never copy to clipboard)
	if raw {
		fmt.Print(entry.Value)
	} else {
		fmt.Println(entry.Value)
	}
}

func handleRm(args []string) {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	_ = fs.Parse(args)

	keys := fs.Args()
	if len(keys) == 0 {
		fmt.Fprintln(os.Stderr, "rm requires at least one key")
		os.Exit(exitError)
	}
	s := loadStore()
	for _, k := range keys {
		if s.Delete(k) {
			// ok
		} else {
			fmt.Fprintf(os.Stderr, "warning: key %q not present\n", k)
		}
	}
	saveStore(s)
}

func handleLs(args []string) {
	fs := flag.NewFlagSet("ls", flag.ExitOnError)
	var all bool
	var long bool
	fs.BoolVar(&all, "all", false, "include expired entries")
	fs.BoolVar(&long, "long", false, "show timestamps")
	_ = fs.Parse(args)

	s := loadStore()
	list := s.List(all)

	if len(list) == 0 {
		fmt.Println("(no entries)")
		return
	}
	now := time.Now()
	for k, e := range list {
		expired := e.ExpiresAt != nil && now.After(*e.ExpiresAt)
		marker := ""
		if expired {
			marker = "(expired)"
		}
		if long {
			expStr := "never"
			if e.ExpiresAt != nil {
				expStr = e.ExpiresAt.Format(time.RFC3339)
			}
			fmt.Printf("%s %s - created:%s expires:%s\n", k, marker, e.CreatedAt.Format(time.RFC3339), expStr)
		} else {
			if marker != "" {
				fmt.Printf("%s %s\n", k, marker)
			} else {
				fmt.Println(k)
			}
		}
	}
}

func handleClear(args []string) {
	fs := flag.NewFlagSet("clear", flag.ExitOnError)
	var expiredOnly bool
	var all bool
	fs.BoolVar(&expiredOnly, "expired", false, "remove only expired entries")
	fs.BoolVar(&all, "all", false, "wipe the whole store")
	_ = fs.Parse(args)

	if !(expiredOnly || all) {
		fmt.Fprintln(os.Stderr, "clear requires --expired or --all")
		os.Exit(exitError)
	}
	s := loadStore()
	switch {
	case all:
		s.Entries = map[string]store.Entry{}
	case expiredOnly:
		removed := s.PurgeExpired()
		fmt.Printf("removed %d expired entrie(s)\n", removed)
	}
	saveStore(s)
}

func handleCp(args []string) {
	// cp is just get with default copy behaviour
	handleGet(args)
}

func handlePut(args []string) {
	// put <key> [--ttl ...] [--temp]
	fs := flag.NewFlagSet("put", flag.ExitOnError)
	var ttlStr string
	var temp bool
	fs.StringVar(&ttlStr, "ttl", "", "time‑to‑live")
	fs.BoolVar(&temp, "temp", false, "temporary entry (default ttl)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "put requires a key")
		os.Exit(exitError)
	}
	key := rest[0]

	// reuse set logic but force --from-stdin
	setArgs := []string{key, "--from-stdin"}
	if ttlStr != "" {
		setArgs = append(setArgs, "--ttl", ttlStr)
	}
	if temp {
		setArgs = append(setArgs, "--temp")
	}
	handleSet(setArgs)
}

func handleMv(args []string) {
	fs := flag.NewFlagSet("mv", flag.ExitOnError)
	var force bool
	fs.BoolVar(&force, "force", false, "overwrite destination if exists")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) != 2 {
		fmt.Fprintln(os.Stderr, "mv requires <old> <new>")
		os.Exit(exitError)
	}
	oldKey, newKey := rest[0], rest[1]

	s := loadStore()
	entry, ok := s.Get(oldKey)
	if !ok {
		fmt.Fprintf(os.Stderr, "source key %q not found or expired\n", oldKey)
		os.Exit(exitNotFound)
	}
	if _, exists := s.Entries[newKey]; exists && !force {
		fmt.Fprintf(os.Stderr, "destination key %q already exists (use --force)\n", newKey)
		os.Exit(exitError)
	}
	s.Set(newKey, entry)
	s.Delete(oldKey)
	saveStore(s)
}

func handleExport() {
	path := getStorePath()
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "failed to read store:", err)
		os.Exit(exitError)
	}
	if len(data) == 0 {
		data = []byte(`{"version":1,"entries":{}}`)
	}
	fmt.Println(string(data))
}

func handleImport() {
	path := getStorePath()
	existing, _ := store.Load(path) // ignore error – will create new if missing
	incomingBytes, err := io.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read stdin:", err)
		os.Exit(exitError)
	}
	var incoming store.Store
	if err := json.Unmarshal(incomingBytes, &incoming); err != nil {
		fmt.Fprintln(os.Stderr, "invalid JSON on stdin:", err)
		os.Exit(exitError)
	}
	for k, v := range incoming.Entries {
		existing.Entries[k] = v // naive merge; later can add --force handling
	}
	if err := store.Save(path, existing); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write store:", err)
		os.Exit(exitError)
	}
}

func handleConfig(args []string) {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	var showPath bool
	fs.BoolVar(&showPath, "path", false, "print the store path")
	_ = fs.Parse(args)

	if showPath {
		fmt.Println(getStorePath())
	}
}

func handleStats() {
	s := loadStore()
	now := time.Now()
	active := 0
	expired := 0
	for _, e := range s.Entries {
		if store.IsExpired(e, now) {
			expired++
		} else {
			active++
		}
	}
	fmt.Printf("active: %d, expired: %d\n", active, expired)
}

/*** End of subcommands ***/
