package main

import (
	"reflect"
	"testing"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestGetConversationState(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer done()

	scs := SerialisedConversationState{
		State: stateNone,
		Data:  `{"hello":"world"}`,
	}

	key := datastore.NewKey(ctx, conversationStateKind, "1:1", 0, nil)
	_, err = datastore.Put(ctx, key, &scs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	state, err := GetConversationState(ctx, 1, 1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expected := ConversationState{
		State: stateNone,
		Data: map[string]string{
			"hello": "world",
		},
	}

	if !reflect.DeepEqual(state, expected) {
		t.Fail()
	}
}

func TestSetConversationState(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer done()

	state := ConversationState{
		State: stateNone,
		Data: map[string]string{
			"hello": "world",
		},
	}

	err = SetConversationState(ctx, 1, 1, state)
	if err != nil {
		t.Fatalf("%v", err)
	}

	key := datastore.NewKey(ctx, conversationStateKind, "1:1", 0, nil)
	var scs SerialisedConversationState
	err = datastore.Get(ctx, key, &scs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expected := SerialisedConversationState{
		State: stateNone,
		Data:  `{"hello":"world"}`,
	}

	if !reflect.DeepEqual(scs, expected) {
		t.Fail()
	}
}
