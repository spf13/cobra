package cobra

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// NamedFlagSets stores named flag sets in the order of calling FlagSet.
type NamedFlagSets struct {
	name          string
	errorHandling pflag.ErrorHandling

	// order is an ordered list of flag set names.
	order []string
	// flagSets stores the flag sets by name.
	flagSets map[string]*pflag.FlagSet
}

func NewNamedFlagSets(name string, errorHandling pflag.ErrorHandling) *NamedFlagSets {
	return &NamedFlagSets{
		name:          name,
		errorHandling: errorHandling,
	}
}

// FlagSet returns the flag set with the given name and adds it to the
// ordered name list if it is not in there yet.
func (nfs *NamedFlagSets) FlagSet(name string) (*pflag.FlagSet, bool) {
	if nfs.flagSets == nil {
		nfs.flagSets = map[string]*pflag.FlagSet{}
	}
	var ok bool
	if _, ok = nfs.flagSets[name]; !ok {
		flagSet := pflag.NewFlagSet(name, nfs.errorHandling)
		nfs.flagSets[name] = flagSet
		nfs.order = append(nfs.order, name)
	}
	return nfs.flagSets[name], ok
}

// Flatten returns a single flag set containing all the flag sets
// in the NamedFlagSet
func (nfs *NamedFlagSets) Flatten() *pflag.FlagSet {
	out := pflag.NewFlagSet(nfs.name, nfs.errorHandling)
	for _, fs := range nfs.flagSets {
		out.AddFlagSet(fs)
	}
	return out
}

// FlagUsages returns a string containing the usage information for all flags in
// the FlagSet
func (nfs *NamedFlagSets) FlagUsages() string {
	return nfs.FlagUsagesWrapped(0)
}

func (nfs *NamedFlagSets) FlagUsagesWrapped(cols int) string {
	var buf bytes.Buffer
	for _, name := range nfs.order {
		fs := nfs.flagSets[name]
		if !fs.HasFlags() {
			continue
		}

		wideFS := pflag.NewFlagSet("", pflag.ExitOnError)
		wideFS.AddFlagSet(fs)

		var zzz string
		if cols > 24 {
			zzz = strings.Repeat("z", cols-24)
			wideFS.Int(zzz, 0, strings.Repeat("z", cols-24))
		}

		s := fmt.Sprintf("\n%s Flags:\n%s", strings.ToUpper(name[:1])+name[1:], wideFS.FlagUsagesWrapped(cols))

		if cols > 24 {
			i := strings.Index(s, zzz)
			lines := strings.Split(s[:i], "\n")
			fmt.Fprint(&buf, strings.Join(lines[:len(lines)-1], "\n"))
			fmt.Fprintln(&buf)
		} else {
			fmt.Fprint(&buf, s)
		}
	}
	return buf.String()
}
