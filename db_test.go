// +build go1.7

package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
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

func NewContext(opts *aetest.Options) (context.Context, func(), error) {
	inst, err := aetest.NewInstance(opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil, nil, err
	}
	ctx := appengine.NewContext(req)
	return ctx, func() {
		inst.Close()
	}, nil
}

func TestNewPack(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	pack1 := Pack{
		Name:    "pack1",
		Creator: 1,
	}

	key := datastore.NewKey(ctx, packKind, strings.ToUpper(pack1.Name), 0, nil)
	_, err = datastore.Put(ctx, key, &pack1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	t.Run("name taken", func(t *testing.T) {
		ok, err := NewPack(ctx, pack1.Name, 2)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if ok {
			t.Fail()
		}

		// check that pack1 is untouched
		pack, err := GetPack(ctx, pack1.Name)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !reflect.DeepEqual(pack, pack1) {
			t.Fail()
		}
	})
	t.Run("ok", func(t *testing.T) {
		ok, err := NewPack(ctx, "pack2", 2)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !ok {
			t.Fail()
		}

		pack2, err := GetPack(ctx, "pack2")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if pack2.Name != "pack2" || pack2.Creator != 2 || len(pack2.Contributors) != 0 {
			t.Fail()
		}
	})
}

func TestMySubscriptions(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewContext(&aetest.Options{
		StronglyConsistentDatastore: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	packs := []Pack{
		{
			Name:    "PACK1",
			Creator: 1,
		},
		{
			Name:    "PACK2",
			Creator: 1,
			Deleted: true,
		},
		{
			Name:    "PACK3",
			Creator: 1,
		},
	}

	var subs []Subscription
	for _, pack := range packs {
		subs = append(subs, Subscription{
			UserID: 1,
			Pack:   pack.Name,
		})
	}

	for _, pack := range packs {
		key := datastore.NewKey(ctx, packKind, pack.Name, 0, nil)
		_, err := datastore.Put(ctx, key, &pack)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, sub := range subs {
		key := datastore.NewKey(ctx, subscriptionKind, fmt.Sprintf("%d:%s", sub.UserID, sub.Pack), 0, nil)
		_, err := datastore.Put(ctx, key, &sub)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("no subs", func(t *testing.T) {
		subs, err := MySubscriptions(ctx, 0)
		assert.Equal(t, err, nil)
		assert.Len(t, subs, 0)
	})

	t.Run("ok", func(t *testing.T) {
		subs, err := MySubscriptions(ctx, 1)
		assert.Equal(t, err, nil)
		assert.Equal(t, subs, []Subscription{{UserID: 1, Pack: "PACK1"}, {UserID: 1, Pack: "PACK3"}})
	})
}

func TestGetUserPacks(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewContext(&aetest.Options{
		StronglyConsistentDatastore: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	packs := []Pack{
		{
			Name:    "PACK1",
			Creator: 1,
		},
		{
			Name:    "PACK2",
			Creator: 1,
			Deleted: true,
		},
		{
			Name:         "PACK3",
			Creator:      2,
			Contributors: []int{1},
		},
		{
			Name:         "PACK4",
			Creator:      2,
			Contributors: []int{1},
			Deleted:      true,
		},
	}

	for _, pack := range packs {
		key := datastore.NewKey(ctx, packKind, pack.Name, 0, nil)
		_, err := datastore.Put(ctx, key, &pack)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("no packs", func(t *testing.T) {
		userPacks, err := GetUserPacks(ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, userPacks.IsContributor, 0)
		assert.Len(t, userPacks.IsCreator, 0)
	})

	t.Run("ok", func(t *testing.T) {
		userPacks, err := GetUserPacks(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, userPacks,
			UserPacks{
				IsCreator: []Pack{
					{
						Name:    "PACK1",
						Creator: 1,
					},
				},
				IsContributor: []Pack{
					{
						Name:         "PACK3",
						Creator:      2,
						Contributors: []int{1},
					},
				},
			})
	})
}

func TestNewContributor(t *testing.T) {
	t.Parallel()

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

	key := datastore.NewKey(ctx, packKind, strings.ToUpper(pack1.Name), 0, nil)
	_, err = datastore.Put(ctx, key, &pack1)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("not creator", func(t *testing.T) {
		_, err := NewContributor(ctx, "pack1", 2, 3)
		if err != nil {
			if err != ErrNotAllowed {
				t.Fatalf("%v", err)
			}

			return
		}

		t.Fail()
	})
	t.Run("already a contributor", func(t *testing.T) {
		ok, err := NewContributor(ctx, "pack1", 1, 2)
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
		ok, err := NewContributor(ctx, "pack1", 1, 3)
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

func TestDeleteContributor(t *testing.T) {
	t.Parallel()

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

	key := datastore.NewKey(ctx, packKind, strings.ToUpper(pack1.Name), 0, nil)
	_, err = datastore.Put(ctx, key, &pack1)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("not creator", func(t *testing.T) {
		_, err := DeleteContributor(ctx, "pack1", 3, 2)
		if err != nil {
			if err != ErrNotAllowed {
				t.Fatalf("%v", err)
			}

			return
		}

		t.Fail()
	})
	t.Run("not a contributor", func(t *testing.T) {
		ok, err := DeleteContributor(ctx, "pack1", 1, 3)
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
		ok, err := DeleteContributor(ctx, "pack1", 1, 2)
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

func TestSubscribe(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewContext(&aetest.Options{
		StronglyConsistentDatastore: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	pack1 := Pack{
		Name:    "pack1",
		Creator: 1,
	}

	subscription1 := Subscription{
		UserID: 1,
		Pack:   "PACK1",
	}

	key1 := datastore.NewKey(ctx, packKind, strings.ToUpper(pack1.Name), 0, nil)
	key2 := datastore.NewKey(ctx, subscriptionKind, "1:PACK1", 0, nil)

	_, err1 := datastore.Put(ctx, key1, &pack1)
	if err1 != nil {
		t.Fatalf("%v", err1)
	}
	_, err2 := datastore.Put(ctx, key2, &subscription1)
	if err2 != nil {
		t.Fatalf("%v", err2)
	}

	t.Run("nonexistent pack", func(t *testing.T) {
		_, err := Subscribe(ctx, "pack2", 1)
		if err != nil {
			if err != ErrNotFound {
				t.Fatalf("%v", err)
			}

			return
		}

		t.Fail()
	})
	t.Run("already subscribed", func(t *testing.T) {
		ok, err := Subscribe(ctx, pack1.Name, subscription1.UserID)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if ok {
			t.Fail()
		}
	})
	t.Run("ok", func(t *testing.T) {
		ok, err := Subscribe(ctx, pack1.Name, 2)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !ok {
			t.Fail()
		}
	})
}

func TestUnsubscribe(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewContext(&aetest.Options{
		StronglyConsistentDatastore: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	pack1 := Pack{
		Name:    "pack1",
		Creator: 1,
	}

	subscription1 := Subscription{
		UserID: 1,
		Pack:   "PACK1",
	}

	pack2 := Pack{
		Name:    "pack2",
		Creator: 1,
		Deleted: true,
	}

	sub2 := Subscription{
		UserID: 1,
		Pack:   "pack2",
	}

	srcs := []interface{}{
		&pack1,
		&subscription1,
		&pack2,
		&sub2,
	}

	keys := []*datastore.Key{
		datastore.NewKey(ctx, packKind, strings.ToUpper(pack1.Name), 0, nil),
		datastore.NewKey(ctx, subscriptionKind, "1:PACK1", 0, nil),
		datastore.NewKey(ctx, packKind, strings.ToUpper(pack2.Name), 0, nil),
		datastore.NewKey(ctx, subscriptionKind, "1:PACK2", 0, nil),
	}

	for i, key := range keys {
		_, err := datastore.Put(ctx, key, srcs[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("nonexistent pack", func(t *testing.T) {
		_, err := Unsubscribe(ctx, "pack3", 1)
		if err != nil {
			if err != ErrNotFound {
				t.Fatalf("%v", err)
			}

			return
		}

		t.Fail()
	})

	t.Run("not subscribed", func(t *testing.T) {
		ok, err := Unsubscribe(ctx, pack1.Name, 2)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if ok {
			t.Fail()
		}
	})

	t.Run("deleted", func(t *testing.T) {
		_, err := Unsubscribe(ctx, pack2.Name, 1)
		assert.Equal(t, err, ErrDeleted)
	})

	t.Run("ok", func(t *testing.T) {
		ok, err := Unsubscribe(ctx, pack1.Name, subscription1.UserID)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !ok {
			t.Fail()
		}
	})
}

func TestSoftDeletePack(t *testing.T) {
	ctx, done, err := NewContext(&aetest.Options{
		StronglyConsistentDatastore: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	pack1 := Pack{
		Name:    "pack1",
		Creator: 1,
	}

	key := datastore.NewKey(ctx, packKind, strings.ToUpper(pack1.Name), 0, nil)
	_, err = datastore.Put(ctx, key, &pack1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	t.Run("invalid name", func(t *testing.T) {
		err := SoftDeletePack(ctx, "!@#$%", 0)
		assert.Equal(t, err, ErrInvalidName)
	})

	t.Run("not creator", func(t *testing.T) {
		err := SoftDeletePack(ctx, pack1.Name, 0)
		assert.Equal(t, err, ErrNotAllowed)
	})

	t.Run("ok", func(t *testing.T) {
		err := SoftDeletePack(ctx, pack1.Name, 1)
		assert.Equal(t, err, nil)

		pack, err := GetPack(ctx, pack1.Name)
		if err != nil {
			assert.Equal(t, err, ErrDeleted)
		}

		assert.True(t, pack.Deleted)
	})
}
