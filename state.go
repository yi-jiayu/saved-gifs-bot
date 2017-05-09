package main

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const conversationStateKind = "ConversationState"

// ConversationState represents a stored conversation state in datastore
type ConversationState struct {
	State []byte
}

// GetConversationState retrieves the current conversation state for userID in chatID
func GetConversationState(ctx context.Context, chatID int64, userID int) (map[string]string, error) {
	key := datastore.NewKey(ctx, conversationStateKind, fmt.Sprintf("%d:%d", chatID, userID), 0, nil)
	var s ConversationState
	err := datastore.Get(ctx, key, &s)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}

		return nil, err
	}

	var state map[string]string
	b := bytes.NewBuffer(s.State)
	d := gob.NewDecoder(b)
	err = d.Decode(&state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// SetConversationState sets the conversation state for userID in chatID.
func SetConversationState(ctx context.Context, chatID int64, userID int, state map[string]string) error {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	err := e.Encode(state)
	if err != nil {
		return err
	}

	s := ConversationState{
		State: b.Bytes(),
	}

	key := datastore.NewKey(ctx, conversationStateKind, fmt.Sprintf("%d:%d", chatID, userID), 0, nil)
	_, err = datastore.Put(ctx, key, &s)
	if err != nil {
		return err
	}

	return nil
}
