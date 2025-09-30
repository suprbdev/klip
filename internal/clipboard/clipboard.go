package clipboard

import (
	"os"

	"github.com/atotto/clipboard"
)

const envDisable = "KLIP_DISABLE_CLIPBOARD"

// Enabled returns true unless the environment variable disables it.
func Enabled() bool {
	if os.Getenv(envDisable) == "1" || os.Getenv(envDisable) == "true" {
		return false
	}
	return true
}

// Copy copies text to the system clipboard if enabled.
// Returns an error from the underlying library, or nil on success.
func Copy(text string) error {
	if !Enabled() {
		return nil // silent no‑op when disabled globally
	}
	return clipboard.WriteAll(text)
}
