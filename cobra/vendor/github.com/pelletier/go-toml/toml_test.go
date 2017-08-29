// Testing support for go-toml

package toml

import (
	"testing"
)

func TestTomlHas(t *testing.T) {
	tree, _ := Load(`
		[test]
		key = "value"
	`)

	if !tree.Has("test.key") {
		t.Errorf("Has - expected test.key to exists")
	}

	if tree.Has("") {
		t.Errorf("Should return false if the key is not provided")
	}
}

func TestTomlGet(t *testing.T) {
	tree, _ := Load(`
		[test]
		key = "value"
	`)

	if tree.Get("") != tree {
		t.Errorf("Get should return the tree itself when given an empty path")
	}

	if tree.Get("test.key") != "value" {
		t.Errorf("Get should return the value")
	}
	if tree.Get(`\`) != nil {
		t.Errorf("should return nil when the key is malformed")
	}
}

func TestTomlGetDefault(t *testing.T) {
	tree, _ := Load(`
		[test]
		key = "value"
	`)

	if tree.GetDefault("", "hello") != tree {
		t.Error("GetDefault should return the tree itself when given an empty path")
	}

	if tree.GetDefault("test.key", "hello") != "value" {
		t.Error("Get should return the value")
	}

	if tree.GetDefault("whatever", "hello") != "hello" {
		t.Error("GetDefault should return the default value if the key does not exist")
	}
}

func TestTomlHasPath(t *testing.T) {
	tree, _ := Load(`
		[test]
		key = "value"
	`)

	if !tree.HasPath([]string{"test", "key"}) {
		t.Errorf("HasPath - expected test.key to exists")
	}
}

func TestTomlGetPath(t *testing.T) {
	node := newTree()
	//TODO: set other node data

	for idx, item := range []struct {
		Path     []string
		Expected *Tree
	}{
		{ // empty path test
			[]string{},
			node,
		},
	} {
		result := node.GetPath(item.Path)
		if result != item.Expected {
			t.Errorf("GetPath[%d] %v - expected %v, got %v instead.", idx, item.Path, item.Expected, result)
		}
	}

	tree, _ := Load("[foo.bar]\na=1\nb=2\n[baz.foo]\na=3\nb=4\n[gorf.foo]\na=5\nb=6")
	if tree.GetPath([]string{"whatever"}) != nil {
		t.Error("GetPath should return nil when the key does not exist")
	}
}

func TestTomlFromMap(t *testing.T) {
	simpleMap := map[string]interface{}{"hello": 42}
	tree, err := TreeFromMap(simpleMap)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if tree.Get("hello") != int64(42) {
		t.Fatal("hello should be 42, not", tree.Get("hello"))
	}
}
