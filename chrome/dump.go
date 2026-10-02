package chrome

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// KeySave is the conventional binding for saving the current view to a file
// (k9s uses ctrl+s for its screen dumps). Like [KeyToggleCrumbs], it is a
// name: apps match it in their own Update, render the view as text, and
// hand it to [SaveDump].
const KeySave = "ctrl+s"

// dumpTimeLayout stamps a dump's filename; it sorts in time order.
const dumpTimeLayout = "20060102-150405"

// defaultDumpName stands in for a name with nothing filename-safe in it.
const defaultDumpName = "dump"

// SaveDump writes content to a new file in dir and returns its path, for a
// status flash such as "saved to <path>". ANSI escapes are stripped, so a
// styled table or log view lands as plain text, and the file ends in a
// newline. dir is created if it is missing.
//
// The file is named <name>-<timestamp>.txt, with name lowercased and every
// character outside [a-z0-9._-] replaced by "-", so a view title such as
// "Pods(all)" is a safe filename. A second save within the same second gets
// a -2, -3, … suffix rather than overwriting the first.
//
// Render the content from the data rather than the screen when the view
// scrolls: table.PlainText gives every row of a table, and a tail's
// VisibleLines joined with "\n" gives every line that passes its filter.
func SaveDump(dir, name, content string) (string, error) {
	return saveDump(dir, name, content, time.Now())
}

func saveDump(dir, name, content string, now time.Time) (string, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create dump dir: %w", err)
	}
	text := ansi.Strip(content)
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	base := dumpName(name) + "-" + now.Format(dumpTimeLayout)
	for n := 1; ; n++ {
		file := base
		if n > 1 {
			file = fmt.Sprintf("%s-%d", base, n)
		}
		path := filepath.Join(dir, file+".txt")
		f, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("create dump: %w", err)
		}
		_, err = f.WriteString(text)
		if err = errors.Join(err, f.Close()); err != nil {
			return "", fmt.Errorf("write dump: %w", err)
		}
		return path, nil
	}
}

// dumpName makes name safe as a filename stem.
func dumpName(name string) string {
	stem := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			return r
		default:
			return '-'
		}
	}, strings.ToLower(name))
	stem = strings.Trim(stem, "-.")
	if stem == "" {
		return defaultDumpName
	}
	return stem
}
