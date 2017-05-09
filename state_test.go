package main

import (
	"reflect"
	"testing"

	"google.golang.org/appengine/aetest"
)

func TestGetSetConversationState(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	state := map[string]string{
		"key": "value",
	}

	err = SetConversationState(ctx, 0, 0, state)
	if err != nil {
		t.Fatalf("err in set: %v", err)
	}

	state2, err := GetConversationState(ctx, 0, 0)
	if err != nil {
		t.Fatalf("err in get: %v", err)
	}

	eq := reflect.DeepEqual(state, state2)
	if !eq {
		t.Fatal("inconsistent state")
	}
}
