package florilegium_test

import (
	"io"
	"os"
	"testing"

	"blekksprut.net/florilegium"
)

const source = ".florilegium-testing"

var g florilegium.Garden

func TestMain(m *testing.M) {
	err := os.Mkdir(source, 0750)
	if err != nil {
		panic(err)
	}
	root, err := os.OpenRoot(source)
	if err != nil {
		panic(err)
	}
	g.Root = root
	code := m.Run()
	err = os.RemoveAll(source)
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestPlant(t *testing.T) {
	err := g.Plant("home", "a home for all")
	if err != nil {
		t.Error("planting home shouldn't fail", err)
	}
}

func TestRead(t *testing.T) {
	err := g.Read("home", io.Discard)
	if err != nil {
		t.Error("reading home shouldn't fail", err)
	}
}

func TestPlantInvalidName(t *testing.T) {
	err := g.Plant("home//wow", "a home for all")
	if err == nil {
		t.Error("shouldn't allow invalid names")
	}
}

func TestReadInvalidName(t *testing.T) {
	err := g.Read("home//wow", io.Discard)
	if err == nil {
		t.Error("shouldn't allow invalid names")
	}
}

func TestReadNotFound(t *testing.T) {
	err := g.Read("notfound", io.Discard)
	if err == nil {
		t.Error("should fail when something isn't planted")
	}
}

func TestStroll(t *testing.T) {
	count := 0
	g.Stroll(func(path string) {
		count += 1
	})
	if count != 1 {
		t.Error("uh-oh garden is overgrown, or something is missing:", count)
	}
}

// coverage
func TestRaw(t *testing.T) {
	g.Raw("home", io.Discard)
}

func TestRawInvalidName(t *testing.T) {
	g.Raw(".", io.Discard)
}

func TestRawNotFound(t *testing.T) {
	g.Raw("notfound", io.Discard)
}
