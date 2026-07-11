//go:build go1.18
// +build go1.18

package cobra

import (
	"strings"
	"testing"
)

// FuzzLd fuzzes the Levenshtein distance implementation and checks basic invariants.
func FuzzLd(f *testing.F) {
	// Seed corpus
	f.Add("", "")
	f.Add("a", "")
	f.Add("", "a")
	f.Add("kitten", "sitting")
	f.Add("Saturday", "Sunday")
	f.Add("MixedCase", "mixedcase")

	f.Fuzz(func(t *testing.T, a string, b string) {
		// Distance is always >= 0
		d := ld(a, b, false)
		if d < 0 {
			t.Fatalf("ld returned negative distance: %d", d)
		}

		// Symmetry: ld(a,b) == ld(b,a)
		d2 := ld(b, a, false)
		if d != d2 {
			t.Fatalf("ld not symmetric: ld(%q,%q)=%d ld(%q,%q)=%d", a, b, d, b, a, d2)
		}

		// Case-insensitive should be <= case-sensitive for inputs differing only by case
		if strings.EqualFold(a, b) {
			di := ld(a, b, true)
			if di != 0 {
				t.Fatalf("ld case-insensitive mismatch for equalFold inputs: %q %q => %d", a, b, di)
			}
		}

		// Identity: distance is 0 when strings are equal
		if a == b && d != 0 {
			t.Fatalf("ld identity broken: ld(%q,%q)=%d", a, b, d)
		}
	})
}

// FuzzConfigEnvVar fuzzes configEnvVar to ensure it only produces A-Z0-9_ and is stable for allowed inputs.
func FuzzConfigEnvVar(f *testing.F) {
	f.Add("prog", "ACTIVE_HELP")
	f.Add("My-App", "COMPLETION_DESCRIPTIONS")
	f.Add("", "X")
	f.Fuzz(func(t *testing.T, name, suffix string) {
		v := configEnvVar(name, suffix)
		if v == "" {
			t.Fatal("empty env var name")
		}
		// Must contain only A-Z0-9_
		for i := 0; i < len(v); i++ {
			c := v[i]
			if !(c == '_' || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
				t.Fatalf("invalid char %q in %q", c, v)
			}
		}
		// Ensure uppercasing behavior for simple alnum inputs
		cleanName := name
		for i := 0; i < len(cleanName); i++ {
			b := cleanName[i]
			if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')) {
				cleanName = ""
				break
			}
		}
		if cleanName != "" {
			upper := strings.ToUpper(cleanName + "_" + suffix)
			// The sanitizer replaces non A-Z0-9_ with _, which shouldn't occur here.
			if v != strings.Map(func(r rune) rune {
				if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
					return r
				}
				return '_'
			}, upper) {
				t.Fatalf("unexpected mapping for simple input: got %q want %q", v, upper)
			}
		}
	})
}
