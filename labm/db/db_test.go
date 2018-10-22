package db

import (
	"testing"
)

func TestDbCrud(t *testing.T) {
	store := NewDb()
	defer store.Close()
	if err := store.Set("u1", "v1"); err != nil {
		t.Fatal(err)
	}

	val := "v1"
	res, err := store.Get("u1")
	if err != nil {
		t.Fatal(err)
	}

	if res != val {
		t.Fatalf("got wrong value %s : %s", val, res)
	}
}

func TestDbRead(t *testing.T) {
	store := NewDb()
	defer store.Close()
	val := "v1"
	res, err := store.Get("u12")
	if err != nil {
		t.Fatal(err)
	}
	if res != val {
		t.Fatalf("got wrong value %s : %s", val, res)
	}
}
