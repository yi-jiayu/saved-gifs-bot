package main

import (
	"reflect"
	"testing"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func sliceSameContents(s1, s2 []int) bool {
	// convert to map
	m1 := make(map[int]struct{})
	for _, v := range s1 {
		m1[v] = struct{}{}
	}

	m2 := make(map[int]struct{})
	for _, v := range s2 {
		m2[v] = struct{}{}
	}

	return reflect.DeepEqual(m1, m2)
}

func TestAddContributor(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	pack1 := Pack{
		Name:         "pack1",
		Creator:      1,
		Contributors: []int{2},
	}

	key := datastore.NewKey(ctx, packKind, pack1.Name, 0, nil)
	_, err = datastore.Put(ctx, key, &pack1)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("not creator", func(t *testing.T) {
		_, err := AddContributor(ctx, "pack1", 2, 3)
		if err != nil {
			if err != ErrNotAllowed {
				t.Fatalf("%v", err)
			}

			return
		}

		t.Fail()
	})
	t.Run("already a contributor", func(t *testing.T) {
		ok, err := AddContributor(ctx, "pack1", 1, 2)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if ok {
			t.Fail()
		}

		pack, err := GetPack(ctx, "pack1")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !sliceSameContents([]int{2}, pack.Contributors) {
			t.Fail()
		}
	})
	t.Run("ok", func(t *testing.T) {
		ok, err := AddContributor(ctx, "pack1", 1, 3)
		if err != nil {
			t.Fatalf("%v", err)
			return
		}

		if !ok {
			t.Fail()
		}

		pack, err := GetPack(ctx, "pack1")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !sliceSameContents([]int{2, 3}, pack.Contributors) {
			t.Fail()
		}
	})
}

func TestRemoveContributor(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	pack1 := Pack{
		Name:         "pack1",
		Creator:      1,
		Contributors: []int{2},
	}

	key := datastore.NewKey(ctx, packKind, pack1.Name, 0, nil)
	_, err = datastore.Put(ctx, key, &pack1)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("not creator", func(t *testing.T) {
		_, err := RemoveContributor(ctx, "pack1", 3, 2)
		if err != nil {
			if err != ErrNotAllowed {
				t.Fatalf("%v", err)
			}

			return
		}

		t.Fail()
	})
	t.Run("not a contributor", func(t *testing.T) {
		ok, err := RemoveContributor(ctx, "pack1", 1, 3)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if ok {
			t.Fail()
		}

		pack, err := GetPack(ctx, "pack1")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !sliceSameContents([]int{2}, pack.Contributors) {
			t.Fail()
		}
	})
	t.Run("ok", func(t *testing.T) {
		ok, err := RemoveContributor(ctx, "pack1", 1, 2)
		if err != nil {
			t.Fatalf("%v", err)
			return
		}

		if !ok {
			t.Fail()
		}

		pack, err := GetPack(ctx, "pack1")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !sliceSameContents([]int{}, pack.Contributors) {
			t.Fail()
		}
	})
}
