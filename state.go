package saved_gifs_bot

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const conversationStateKind = "ConversationState"

type ConversationState struct {
	State []byte
}

func GetConversationState(ctx context.Context, chatId int64, userId int) (map[string]string, error) {
	key := datastore.NewKey(ctx, conversationStateKind, fmt.Sprintf("%d:%d", chatId, userId), 0, nil)
	var s ConversationState
	err := datastore.Get(ctx, key, &s)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		} else {
			return nil, err
		}
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

func SetConversationState(ctx context.Context, chatId int64, userId int, state map[string]string) error {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	err := e.Encode(state)
	if err != nil {
		return err
	}

	s := ConversationState{
		State: b.Bytes(),
	}

	key := datastore.NewKey(ctx, conversationStateKind, fmt.Sprintf("%d:%d", chatId, userId), 0, nil)
	_, err = datastore.Put(ctx, key, &s)
	if err != nil {
		return err
	}

	return nil
}
