package chrome

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestKeySaveMatchesBubbletea(t *testing.T) {
	t.Parallel()
	if got := (tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}).String(); got != KeySave {
		t.Errorf("KeySave = %q, bubbletea says %q", KeySave, got)
	}
}

func TestSaveDump_WritesPlainTextAndReturnsPath(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "screen-dumps") // created on demand
	now := time.Date(2026, 10, 2, 14, 5, 9, 0, time.UTC)
	path, err := saveDump(dir, "Pods(all)", "\x1b[1;36mNAME\x1b[m  AGE\nweb   5m", now)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "pods-all-20261002-140509.txt"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "NAME  AGE\nweb   5m\n"; string(got) != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestSaveDump_SameSecondDoesNotOverwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	first, err := saveDump(dir, "logs", "one\n", now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := saveDump(dir, "logs", "two\n", now)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasSuffix(second, "-2.txt") {
		t.Fatalf("second save = %q (first %q)", second, first)
	}
	if b, _ := os.ReadFile(first); string(b) != "one\n" {
		t.Errorf("first dump overwritten: %q", b)
	}
}

func TestSaveDump_PublicUsesNow(t *testing.T) {
	t.Parallel()
	path, err := SaveDump(t.TempDir(), "x", "y")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(filepath.Base(path), "x-"+time.Now().Format("20060102")) {
		t.Errorf("unexpected name %q", filepath.Base(path))
	}
}

func TestSaveDump_DirError(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveDump(filepath.Join(file, "sub"), "x", "y"); err == nil {
		t.Error("a dir under a regular file should fail")
	}
}

func TestDumpName(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"Pods(all)":        "pods-all",
		"../../etc/passwd": "etc-passwd",
		"logs web-1":       "logs-web-1",
		"":                 "dump",
		"///":              "dump",
		"v1.2_x":           "v1.2_x",
	} {
		if got := dumpName(in); got != want {
			t.Errorf("dumpName(%q) = %q, want %q", in, got, want)
		}
	}
}
