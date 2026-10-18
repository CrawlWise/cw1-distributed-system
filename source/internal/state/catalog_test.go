package state

import (
	"strings"
	"testing"

	"example.edu/6cc545/cw1/internal/model"
)

func TestWeek03_Catalog(t *testing.T) {
	c := NewCatalog()
	a := []model.FileMetadata{{Name: "b.txt", Size: 0, SHA256: strings.Repeat("a", 64)}, {Name: "a.txt", Size: 1, SHA256: strings.Repeat("b", 64)}}
	if e := c.Replace(a); e != nil {
		t.Fatal(e)
	}
	a[0].Name = "mutated"
	got := c.Snapshot()
	if len(got) != 2 || got[0].Name != "a.txt" || got[1].Name != "b.txt" {
		t.Fatal("catalogue must copy and order input")
	}
	got[0].Name = "mutated"
	if c.Snapshot()[0].Name != "a.txt" {
		t.Fatal("snapshot aliases state")
	}
	bad := []model.FileMetadata{{Name: "../private", Size: 1, SHA256: "x"}}
	if c.Replace(bad) == nil {
		t.Fatal("invalid metadata accepted")
	}
	if c.Snapshot()[0].Name != "a.txt" {
		t.Fatal("rejected replacement changed state")
	}
}

func TestWeek03_CatalogDuplicateNames(t *testing.T) {
	c := NewCatalog()

	good := model.FileMetadata{
		Name:   "notes.txt",
		Size:   10,
		SHA256: strings.Repeat("a", 64),
	}

	if err := c.Replace([]model.FileMetadata{good}); err != nil {
		t.Fatal(err)
	}

	duplicate := []model.FileMetadata{
		good,
		good,
	}

	if err := c.Replace(duplicate); err == nil {
		t.Fatal("expected duplicate filename error")
	}

	got := c.Snapshot()

	if len(got) != 1 || got[0].Name != "notes.txt" {
		t.Fatal("duplicate replacement modified existing state")
	}
}

func TestWeek03_CatalogEmpty(t *testing.T) {
	c := NewCatalog()

	got := c.Snapshot()

	if got == nil {
		t.Fatal("expected an empty slice, not nil")
	}

	if len(got) != 0 {
		t.Fatal("new catalogue should be empty")
	}
}
