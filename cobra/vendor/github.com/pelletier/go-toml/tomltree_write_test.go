package toml

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

type failingWriter struct {
	failAt  int
	written int
	buffer  bytes.Buffer
}

func (f failingWriter) Write(p []byte) (n int, err error) {
	count := len(p)
	toWrite := f.failAt - count + f.written
	if toWrite < 0 {
		toWrite = 0
	}
	if toWrite > count {
		f.written += count
		f.buffer.WriteString(string(p))
		return count, nil
	}

	f.buffer.WriteString(string(p[:toWrite]))
	f.written = f.failAt
	return f.written, fmt.Errorf("failingWriter failed after writting %d bytes", f.written)
}

func assertErrorString(t *testing.T, expected string, err error) {
	expectedErr := errors.New(expected)
	if err.Error() != expectedErr.Error() {
		t.Errorf("expecting error %s, but got %s instead", expected, err)
	}
}

func TestTreeWriteToEmptyTable(t *testing.T) {
	doc := `[[empty-tables]]
[[empty-tables]]`

	toml, err := Load(doc)
	if err != nil {
		t.Fatal("Unexpected Load error:", err)
	}
	tomlString, err := toml.ToTomlString()
	if err != nil {
		t.Fatal("Unexpected ToTomlString error:", err)
	}

	expected := `
[[empty-tables]]

[[empty-tables]]
`

	if tomlString != expected {
		t.Fatalf("Expected:\n%s\nGot:\n%s", expected, tomlString)
	}
}

func TestTreeWriteToTomlString(t *testing.T) {
	toml, err := Load(`name = { first = "Tom", last = "Preston-Werner" }
points = { x = 1, y = 2 }`)

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	tomlString, _ := toml.ToTomlString()
	reparsedTree, err := Load(tomlString)

	assertTree(t, reparsedTree, err, map[string]interface{}{
		"name": map[string]interface{}{
			"first": "Tom",
			"last":  "Preston-Werner",
		},
		"points": map[string]interface{}{
			"x": int64(1),
			"y": int64(2),
		},
	})
}

func TestTreeWriteToTomlStringSimple(t *testing.T) {
	tree, err := Load("[foo]\n\n[[foo.bar]]\na = 42\n\n[[foo.bar]]\na = 69\n")
	if err != nil {
		t.Errorf("Test failed to parse: %v", err)
		return
	}
	result, err := tree.ToTomlString()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expected := "\n[foo]\n\n  [[foo.bar]]\n    a = 42\n\n  [[foo.bar]]\n    a = 69\n"
	if result != expected {
		t.Errorf("Expected got '%s', expected '%s'", result, expected)
	}
}

func TestTreeWriteToTomlStringKeysOrders(t *testing.T) {
	for i := 0; i < 100; i++ {
		tree, _ := Load(`
		foobar = true
		bar = "baz"
		foo = 1
		[qux]
		  foo = 1
		  bar = "baz2"`)

		stringRepr, _ := tree.ToTomlString()

		t.Log("Intermediate string representation:")
		t.Log(stringRepr)

		r := strings.NewReader(stringRepr)
		toml, err := LoadReader(r)

		if err != nil {
			t.Fatal("Unexpected error:", err)
		}

		assertTree(t, toml, err, map[string]interface{}{
			"foobar": true,
			"bar":    "baz",
			"foo":    1,
			"qux": map[string]interface{}{
				"foo": 1,
				"bar": "baz2",
			},
		})
	}
}

func testMaps(t *testing.T, actual, expected map[string]interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Fatal("trees aren't equal.\n", "Expected:\n", expected, "\nActual:\n", actual)
	}
}

func TestTreeWriteToMapSimple(t *testing.T) {
	tree, _ := Load("a = 42\nb = 17")

	expected := map[string]interface{}{
		"a": int64(42),
		"b": int64(17),
	}

	testMaps(t, tree.ToMap(), expected)
}

func TestTreeWriteToInvalidTreeSimpleValue(t *testing.T) {
	tree := Tree{values: map[string]interface{}{"foo": int8(1)}}
	_, err := tree.ToTomlString()
	assertErrorString(t, "invalid value type at foo: int8", err)
}

func TestTreeWriteToInvalidTreeTomlValue(t *testing.T) {
	tree := Tree{values: map[string]interface{}{"foo": &tomlValue{int8(1), Position{}}}}
	_, err := tree.ToTomlString()
	assertErrorString(t, "unsupported value type int8: 1", err)
}

func TestTreeWriteToInvalidTreeTomlValueArray(t *testing.T) {
	tree := Tree{values: map[string]interface{}{"foo": &tomlValue{[]interface{}{int8(1)}, Position{}}}}
	_, err := tree.ToTomlString()
	assertErrorString(t, "unsupported value type int8: 1", err)
}

func TestTreeWriteToFailingWriterInSimpleValue(t *testing.T) {
	toml, _ := Load(`a = 2`)
	writer := failingWriter{failAt: 0, written: 0}
	_, err := toml.WriteTo(writer)
	assertErrorString(t, "failingWriter failed after writting 0 bytes", err)
}

func TestTreeWriteToFailingWriterInTable(t *testing.T) {
	toml, _ := Load(`
[b]
a = 2`)
	writer := failingWriter{failAt: 2, written: 0}
	_, err := toml.WriteTo(writer)
	assertErrorString(t, "failingWriter failed after writting 2 bytes", err)

	writer = failingWriter{failAt: 13, written: 0}
	_, err = toml.WriteTo(writer)
	assertErrorString(t, "failingWriter failed after writting 13 bytes", err)
}

func TestTreeWriteToFailingWriterInArray(t *testing.T) {
	toml, _ := Load(`
[[b]]
a = 2`)
	writer := failingWriter{failAt: 2, written: 0}
	_, err := toml.WriteTo(writer)
	assertErrorString(t, "failingWriter failed after writting 2 bytes", err)

	writer = failingWriter{failAt: 15, written: 0}
	_, err = toml.WriteTo(writer)
	assertErrorString(t, "failingWriter failed after writting 15 bytes", err)
}

func TestTreeWriteToMapExampleFile(t *testing.T) {
	tree, _ := LoadFile("example.toml")
	expected := map[string]interface{}{
		"title": "TOML Example",
		"owner": map[string]interface{}{
			"name":         "Tom Preston-Werner",
			"organization": "GitHub",
			"bio":          "GitHub Cofounder & CEO\nLikes tater tots and beer.",
			"dob":          time.Date(1979, time.May, 27, 7, 32, 0, 0, time.UTC),
		},
		"database": map[string]interface{}{
			"server":         "192.168.1.1",
			"ports":          []interface{}{int64(8001), int64(8001), int64(8002)},
			"connection_max": int64(5000),
			"enabled":        true,
		},
		"servers": map[string]interface{}{
			"alpha": map[string]interface{}{
				"ip": "10.0.0.1",
				"dc": "eqdc10",
			},
			"beta": map[string]interface{}{
				"ip": "10.0.0.2",
				"dc": "eqdc10",
			},
		},
		"clients": map[string]interface{}{
			"data": []interface{}{
				[]interface{}{"gamma", "delta"},
				[]interface{}{int64(1), int64(2)},
			},
		},
	}
	testMaps(t, tree.ToMap(), expected)
}

func TestTreeWriteToMapWithTablesInMultipleChunks(t *testing.T) {
	tree, _ := Load(`
	[[menu.main]]
        a = "menu 1"
        b = "menu 2"
        [[menu.main]]
        c = "menu 3"
        d = "menu 4"`)
	expected := map[string]interface{}{
		"menu": map[string]interface{}{
			"main": []interface{}{
				map[string]interface{}{"a": "menu 1", "b": "menu 2"},
				map[string]interface{}{"c": "menu 3", "d": "menu 4"},
			},
		},
	}
	treeMap := tree.ToMap()

	testMaps(t, treeMap, expected)
}

func TestTreeWriteToMapWithArrayOfInlineTables(t *testing.T) {
	tree, _ := Load(`
    	[params]
	language_tabs = [
    		{ key = "shell", name = "Shell" },
    		{ key = "ruby", name = "Ruby" },
    		{ key = "python", name = "Python" }
	]`)

	expected := map[string]interface{}{
		"params": map[string]interface{}{
			"language_tabs": []interface{}{
				map[string]interface{}{
					"key":  "shell",
					"name": "Shell",
				},
				map[string]interface{}{
					"key":  "ruby",
					"name": "Ruby",
				},
				map[string]interface{}{
					"key":  "python",
					"name": "Python",
				},
			},
		},
	}

	treeMap := tree.ToMap()
	testMaps(t, treeMap, expected)
}
