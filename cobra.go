// Copyright 2013-2023 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Commands similar to git, go tools and other modern CLI tools
// inspired by go, go-Commander, gh and subcommand

package cobra

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var templateFuncs = template.FuncMap{
	"trim":                    strings.TrimSpace,
	"trimRightSpace":          trimRightSpace,
	"trimTrailingWhitespaces": trimRightSpace,
	"appendIfNotPresent":      appendIfNotPresent,
	"rpad":                    rpad,
	"gt":                      Gt,
	"eq":                      Eq,
}

var initializers []func()
var finalizers []func()

const (
	defaultPrefixMatching   = false
	defaultCommandSorting   = true
	defaultCaseInsensitive  = false
	defaultTraverseRunHooks = false
)

// EnablePrefixMatching allows setting automatic prefix matching. Automatic prefix matching can be a dangerous thing
// to automatically enable in CLI tools.
// Set this to true to enable it.
var EnablePrefixMatching = defaultPrefixMatching

// EnableCommandSorting controls sorting of the slice of commands, which is turned on by default.
// To disable sorting, set it to false.
var EnableCommandSorting = defaultCommandSorting

// EnableCaseInsensitive allows case-insensitive commands names. (case sensitive by default)
var EnableCaseInsensitive = defaultCaseInsensitive

// EnableTraverseRunHooks executes persistent pre-run and post-run hooks from all parents.
// By default this is disabled, which means only the first run hook to be found is executed.
var EnableTraverseRunHooks = defaultTraverseRunHooks

// MousetrapHelpText enables an information splash screen on Windows
// if the CLI is started from explorer.exe.
// To disable the mousetrap, just set this variable to blank string ("").
// Works only on Microsoft Windows.
var MousetrapHelpText = `This is a command line tool.

You need to open cmd.exe and run it from there.
`

// MousetrapDisplayDuration controls how long the MousetrapHelpText message is displayed on Windows
// if the CLI is started from explorer.exe. Set to 0 to wait for the return key to be pressed.
// To disable the mousetrap, just set MousetrapHelpText to blank string ("").
// Works only on Microsoft Windows.
var MousetrapDisplayDuration = 5 * time.Second

// AddTemplateFunc registers a new template function with the given name, making it available for use in the Usage and Help templates.
func AddTemplateFunc(name string, tmplFunc interface{}) {
	templateFuncs[name] = tmplFunc
}

// AddTemplateFuncs adds multiple template functions that are available to Usage and
// Help template generation. It takes a map of template function names to their implementations
// and merges them into the global template function map, allowing these functions to be used
// in usage and help text templates.
func AddTemplateFuncs(tmplFuncs template.FuncMap) {
	for k, v := range tmplFuncs {
		templateFuncs[k] = v
	}
}

// OnInitialize registers functions that will be executed whenever a command's
// Execute method is invoked. These functions are typically used for setup or initialization tasks.
func OnInitialize(y ...func()) {
	initializers = append(initializers, y...)
}

// OnFinalize sets the provided functions to be executed when each command's
// Execute method is completed. The functions are called in the order they are provided.
func OnFinalize(y ...func()) {
	finalizers = append(finalizers, y...)
}

// FIXME Gt is unused by cobra and should be removed in a version 2. It exists only for compatibility with users of cobra.

// Gt compares two types and returns true if the first type is greater than the second. For arrays, channels,
// maps, and slices, it compares their lengths. Ints are compared directly, while strings are first converted to ints
// before comparison. If either conversion fails (e.g., string is not a valid integer), Gt will return false.
//
// Parameters:
//   a - the first value to compare.
//   b - the second value to compare.
//
// Returns:
//   true if a is greater than b, false otherwise.
//
// Errors:
//   None. The function handles invalid string conversions internally and returns false in such cases.
func Gt(a interface{}, b interface{}) bool {
	var left, right int64
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		left = int64(av.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		left = av.Int()
	case reflect.String:
		left, _ = strconv.ParseInt(av.String(), 10, 64)
	}

	bv := reflect.ValueOf(b)

	switch bv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		right = int64(bv.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		right = bv.Int()
	case reflect.String:
		right, _ = strconv.ParseInt(bv.String(), 10, 64)
	}

	return left > right
}

// FIXME Eq is unused by cobra and should be removed in a version 2. It exists only for compatibility with users of cobra.

// Eq takes two types and checks whether they are equal. Supported types are int and string. Unsupported types will panic.
// Parameters:
//   a - the first value to compare
//   b - the second value to compare
// Returns:
//   true if a and b are equal, false otherwise
// Panics:
//   if a or b is of an unsupported type (array, chan, map, slice)
func Eq(a interface{}, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		panic("Eq called on unsupported type")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return av.Int() == bv.Int()
	case reflect.String:
		return av.String() == bv.String()
	}
	return false
}

// TrimRightSpace returns a copy of the input string with trailing white space removed.
func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// FIXME appendIfNotPresent is unused by cobra and should be removed in a version 2. It exists only for compatibility with users of cobra.

// AppendIfNotPresent appends a string to the end of the slice if it's not already present.
//
// Parameters:
//   - s: The slice to append to.
//   - stringToAppend: The string to be appended.
//
// Returns:
//   The updated slice with the string appended, or the original slice unchanged if the string was already present.
func appendIfNotPresent(s, stringToAppend string) string {
	if strings.Contains(s, stringToAppend) {
		return s
	}
	return s + " " + stringToAppend
}

// rpad adds padding to the right of a string.
//
// Parameters:
//   - s: The input string to pad.
//   - padding: The number of spaces to add as padding on the right.
//
// Returns:
//   - The formatted string with right-padding applied.
func rpad(s string, padding int) string {
	formattedString := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(formattedString, s)
}

// tmpl returns a new tmplFunc instance with the provided text.
func tmpl(text string) *tmplFunc {
	return &tmplFunc{
		tmpl: text,
		fn: func(w io.Writer, data interface{}) error {
			t := template.New("top")
			t.Funcs(templateFuncs)
			template.Must(t.Parse(text))
			return t.Execute(w, data)
		},
	}
}

// ld calculates the Levenshtein distance between two strings, optionally ignoring case.
// It returns the minimum number of single-character edits required to change one word into the other.
// Parameters:
//   s - the first string
//   t - the second string
//   ignoreCase - if true, the comparison is case-insensitive
// Returns:
//   The Levenshtein distance between the two strings.
func ld(s, t string, ignoreCase bool) int {
	if ignoreCase {
		s = strings.ToLower(s)
		t = strings.ToLower(t)
	}
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}

	}
	return d[len(s)][len(t)]
}

// stringInSlice checks if a string is present in the given slice of strings.
//
// Parameters:
//   - a: The string to search for.
//   - list: The slice of strings to search within.
//
// Returns:
//   - true if the string is found in the slice, false otherwise.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// CheckErr prints the provided message with the prefix 'Error:' and exits with an error code of 1. If the message is `nil`, it does nothing.
func CheckErr(msg interface{}) {
	if msg != nil {
		fmt.Fprintln(os.Stderr, "Error:", msg)
		os.Exit(1)
	}
}

// WriteStringAndCheck writes a string into a buffer and checks if the error is not nil.
// It takes an io.StringWriter `b` and a string `s`, writes the string to the writer,
// and calls CheckErr with the resulting error, handling any errors that occur during the write operation.
func WriteStringAndCheck(b io.StringWriter, s string) {
	_, err := b.WriteString(s)
	CheckErr(err)
}
