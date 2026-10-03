// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"path/filepath"
	"testing"
)

// FuzzDumpName checks the guarantee SaveDump's doc makes about a view
// title: whatever the name, the stem is a single, non-empty path element
// drawn from [a-z0-9._-], so a dump can never land outside its directory.
func FuzzDumpName(f *testing.F) {
	for _, s := range []string{"", "Pods(all)", "../../etc/passwd", "..", ".", "a/b", `a\b`, "日本", "-.-"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, name string) {
		stem := dumpName(name)
		if stem == "" || stem == "." || stem == ".." || filepath.Base(stem) != stem {
			t.Fatalf("dumpName(%q) = %q, not a single path element", name, stem)
		}
		for _, r := range stem {
			ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-'
			if !ok {
				t.Fatalf("dumpName(%q) = %q, has %q outside [a-z0-9._-]", name, stem, r)
			}
		}
	})
}
