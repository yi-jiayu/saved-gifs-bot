package main

import (
	"errors"
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/search"
)

// states
const (
	stateNone = iota
	stateNewGifWaitPackName
	stateNewGifWaitGif
	stateNewGifWaitKeywords
	stateNewPackWaitPackName
	stateSubscribeWaitPackName
	stateUnsubscribeWaitPackName
	stateDeleteGifWaitPackName
	stateDeleteGifWaitGif
)

// Transducers is a map associating states with their respective Transducer
var Transducers = map[int]Transducer{
	stateNewGifWaitPackName:      stateNewGifWaitPackNameTransducer,
	stateNewGifWaitGif:           newGifWaitGifTransducer,
	stateNewGifWaitKeywords:      newGifWaitKeywordsTransducer,
	stateNewPackWaitPackName:     newPackWaitPackNameTransducer,
	stateSubscribeWaitPackName:   subscibeWaitPackNameTransducer,
	stateUnsubscribeWaitPackName: unsubscribeWaitPackNameTransducer,
	stateDeleteGifWaitPackName:   deleteGifWaitPackNameTransduce,
	stateDeleteGifWaitGif:        deleteGifWaitGifTransducer,
}

// State errors
var (
	ErrInvalidState = errors.New("invalid state")
)

// A Transducer handles an incoming message (input) based on conversation state (current state), returning an action to
// be performed (output) and returning a new conversation state (next state).
type Transducer func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error)

// Transduce continues a conversation based on an incoming message and the current conversation state.
func Transduce(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	userID := message.From.ID

	// get state
	state, err := GetConversationState(ctx, chatID, userID)
	if err != nil {
		return err
	}

	var action func() error
	var nextState ConversationState
	if transducer := Transducers[state.State]; transducer != nil {
		var err error
		nextState, action, err = transducer(ctx, bot, message, state)
		if err != nil {
			return err
		}
	} else {
		// no transducer for state
		return nil
	}

	// update state
	err = SetConversationState(ctx, chatID, userID, nextState)
	if err != nil {
		return err
	}

	// perform action
	err = action()
	if err != nil {
		return err
	}

	return nil
}

func stateNewGifWaitPackNameTransducer(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	userID := message.From.ID

	var nextState ConversationState
	var text string
	if packName := message.Text; packName != "" {
		pack, err := GetPack(ctx, packName)
		if err != nil {
			if err == ErrNotFound {
				text = "Oh no! That pack does not exist. Did you spell it correctly?"
				nextState = state
			} else if err == ErrInvalidName {
				text = "Oh no! That was an invalid pack name. Did you spell it correctly?"
				nextState = state
			} else {
				return state, nil, err
			}
		} else {
			if HasEditPermissions(pack, userID) {
				text = "Please send me the gif you want to add to this pack."
				nextState = ConversationState{
					State: stateNewGifWaitGif,
					Data: map[string]string{
						"packName": packName,
					},
				}
			} else {
				text = "Oops, it seems like you are not the creator of this pack. Only the pack creator can add gifs to a pack."
				nextState = state
			}
		}
	} else {
		text = "Oops! I was waiting for you to send me the name of the gif pack you want to add a new gif to."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
		reply.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  true,
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}

func newGifWaitGifTransducer(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	packName := state.Data["packName"]

	var nextState ConversationState
	var text string
	if document := message.Document; document.MimeType == "video/mp4" {
		_, err := GetGif(ctx, packName, document.FileID)
		if err != nil {
			if err == ErrNotFound {
				text = "Alright, now send me some keywords that describe this gif."
				nextState = ConversationState{
					State: stateNewGifWaitKeywords,
					Data: map[string]string{
						"packName": packName,
						"fileID":   document.FileID,
					},
				}
			} else {
				return state, nil, err
			}
		} else {
			text = "Oops, that gif is already part of this pack. Perhaps you wanted to edit its keywords instead?"
			nextState = state
		}
	} else {
		text = "Oops, I was waiting for you to send me a gif."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
		reply.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  true,
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}

func newGifWaitKeywordsTransducer(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	userID := message.From.ID

	var nextState ConversationState
	var text string
	if keywords := message.Text; keywords != "" {
		packName := state.Data["packName"]
		gif := Gif{
			Pack:     search.Atom(packName),
			FileID:   search.Atom(state.Data["fileID"]),
			Keywords: keywords,
		}

		ok, err := NewGif(ctx, packName, userID, gif)
		if err != nil {
			return state, nil, err
		}

		if ok {
			text = "Great! A new gif has been added to your gif pack."
		} else {
			text = "Oops, that gif is already part of this pack. You can use /editgif to edit it instead."
		}
		nextState = ConversationState{}
	} else {
		text = "Oops, I was waiting for you to send me some keywords for this gif."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if nextState.State != stateNone {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}

func newPackWaitPackNameTransducer(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	userID := message.From.ID

	var nextState ConversationState
	var text string
	if packName := message.Text; packName != "" {
		created, err := NewPack(ctx, packName, userID)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
				nextState = state
			} else {
				return state, nil, err
			}
		} else {
			if created {
				text = "Great! Your gif pack has been created."
				nextState = ConversationState{}
			} else {
				text = "Oh no! That pack name has already been taken."
				nextState = state
			}
		}
	} else {
		text = "Oops! I was waiting for you to send me a name for your new gif pack."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if nextState.State != stateNone {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}

func subscibeWaitPackNameTransducer(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	userID := message.From.ID

	var text string
	var nextState ConversationState
	if packName := message.Text; packName != "" {
		subscribed, err := Subscribe(ctx, packName, userID)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
				nextState = state
			} else if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
				nextState = state
			} else {
				return state, nil, err
			}
		} else {
			if subscribed {
				text = "Great! You have been subscribed to this gif pack!"
			} else {
				text = "Don't worry, you are already subscribed to this gif pack!"
			}
			nextState = ConversationState{}
		}
	} else {
		text = "Oops, I was waiting for you to send me the name of the gif pack you want to subcribe to."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if nextState.State != stateNone {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}

func unsubscribeWaitPackNameTransducer(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	userID := message.From.ID

	var text string
	var nextState ConversationState
	if packName := message.Text; packName != "" {
		unsubscribed, err := Unsubscribe(ctx, packName, userID)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
				nextState = state
			} else if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
				nextState = state
			} else {
				return state, nil, err
			}
		} else {
			if unsubscribed {
				text = "Great! You have been unsubscribed from that gif pack."
				nextState = ConversationState{}
			} else {
				text = "Don't worry, it seems like you were never subscribed to that gif pack in the first place."
				nextState = ConversationState{}
			}
		}
	} else {
		text = "Oops, I was waiting for you to send me the name of the gif pack you want to unsubscribe from."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if nextState.State != stateNone {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}

func deleteGifWaitPackNameTransduce(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	var text string
	var nextState ConversationState
	if packName := message.Text; packName != "" {
		_, err := GetPack(ctx, packName)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
				nextState = state
			} else if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
				nextState = state
			} else {
				return state, nil, err
			}
		} else {
			text = "Please send me the gif you want to delete from this pack."
			state.State = stateDeleteGifWaitGif
			state.Data["packName"] = packName
			nextState = state
		}
	} else {
		text = "Oops, I was waiting for you to send me the name of the gif pack you want to delete a gif from."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if nextState.State != stateNone {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}

func deleteGifWaitGifTransducer(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, state ConversationState) (ConversationState, func() error, error) {
	userID := message.From.ID

	var text string
	var nextState ConversationState
	if document := message.Document; document.MimeType == "video/mp4" {
		packName := state.Data["packName"]
		fileID := document.FileID

		ok, err := DeleteGif(ctx, packName, userID, fileID)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
				nextState = state
			} else if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
				nextState = state
			} else {
				return state, nil, err
			}
		}

		if ok {
			text = "Great, that gif has been deleted!"
			nextState = ConversationState{}
		} else {
			text = "Oops, I couldn't find that gif in this pack. Did you send the right one?"
			nextState = state
		}
	} else {
		text = "Oops, I was waiting for you to send me a gif to be deleted from this pack."
		nextState = state
	}

	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if nextState.State != stateNone {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	action := func() error {
		_, err := bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nextState, action, nil
}
