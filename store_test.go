package main

import (
	"bytes"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "mybestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedFilenameKey := "be17b32c2870b1c0c73b59949db6a3be7814dd23"
	expectedPathname := "be17b/32c28/70b1c/0c73b/59949/db6a3/be781/4dd23"
	if pathKey.Pathname != expectedPathname {
		t.Errorf("have %s want %s", pathKey.Pathname, expectedPathname)
	}
	if pathKey.Filename != expectedFilenameKey {
		t.Errorf("have %s want %s", pathKey.Filename, expectedFilenameKey)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	key := "myspecialpicture"
	data := []byte("some jpg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if ok := s.HasKey(key); !ok {
		t.Errorf("expected to have key %s", key)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func newStore() *Store {
	ops := StoreOps{
		PathTransformFunc: CASPathTransformFunc,
	}

	return NewStore(ops)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
