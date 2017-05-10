package main

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const conversationStateKind = "SerialisedConversationState"

// SerialisedConversationState represents a stored conversation state in datastore
type SerialisedConversationState struct {
	State int
	Data  string
}

// ConversationState represents the state of a conversation
type ConversationState struct {
	State int
	Data  map[string]string
}

// GetConversationState retrieves the current conversation state for userID in chatID
func GetConversationState(ctx context.Context, chatID int64, userID int) (ConversationState, error) {
	key := datastore.NewKey(ctx, conversationStateKind, fmt.Sprintf("%d:%d", chatID, userID), 0, nil)
	var s SerialisedConversationState
	err := datastore.Get(ctx, key, &s)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			// initialise map in case it is assigned to later
			state := ConversationState{
				Data: make(map[string]string),
			}
			return state, nil
		}

		return ConversationState{}, err
	}

	var data map[string]string
	err = json.Unmarshal([]byte(s.Data), &data)
	if err != nil {
		return ConversationState{}, err
	}

	state := ConversationState{
		State: s.State,
		Data:  data,
	}

	return state, nil
}

// SetConversationState sets the conversation state for userID in chatID.
func SetConversationState(ctx context.Context, chatID int64, userID int, state ConversationState) error {
	if state.Data == nil {
		state.Data = make(map[string]string)
	}

	data, err := json.Marshal(state.Data)
	if err != nil {
		return err
	}

	scs := SerialisedConversationState{
		State: state.State,
		Data:  string(data),
	}

	key := datastore.NewKey(ctx, conversationStateKind, fmt.Sprintf("%d:%d", chatID, userID), 0, nil)
	_, err = datastore.Put(ctx, key, &scs)
	if err != nil {
		return err
	}

	return nil
}

// ClearConversationState is a convenience method for clearing the conversation state for user
func ClearConversationState(ctx context.Context, chatID int64, userID int) error {
	key := datastore.NewKey(ctx, conversationStateKind, fmt.Sprintf("%d:%d", chatID, userID), 0, nil)
	err := datastore.Delete(ctx, key)
	if err != nil {
		return err
	}

	return nil
}
