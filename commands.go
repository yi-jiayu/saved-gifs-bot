package main

import (
	"fmt"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
)

var commandHandlers = map[string]MessageHandler{
	"newpack":       cmdNewPackHandler,
	"mypacks":       cmdMyPacksHandler,
	"subscribe":     cmdSubscribeHandler,
	"unsubscribe":   cmdUnsubscribeHandler,
	"subscriptions": cmdSubscriptionsHandler,
	"newgif":        cmdNewGifHandler,
	"deletegif":     cmdDeleteGifHandler,
	"version":       cmdVersionHandler,
	"cancel":        cmdCancelHandler,
}

// MessageHandler represents a function which handles an incoming message.
type MessageHandler func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error

func cmdNewPackHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	userID := message.From.ID
	chatID := message.Chat.ID

	var text string
	done := false
	if name := message.CommandArguments(); name != "" {
		created, err := NewPack(ctx, name, userID)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
				done = true
			} else {
				return err
			}
		} else {
			if created {
				text = "Great! Your gif pack has been created."
				done = true
			} else {
				text = "Oh no! That pack name has already been taken."
				done = true
			}
		}
	} else {
		text = "What do you want to call your new gif pack?"
	}

	if !done {
		state := ConversationState{
			State: stateNewPackWaitPackName,
		}

		err := SetConversationState(ctx, chatID, userID, state)
		if err != nil {
			return err
		}
	}

	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if !done {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func cmdMyPacksHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	userID := message.From.ID
	chatID := message.Chat.ID

	var text string
	userPacks, err := GetUserPacks(ctx, userID)
	if err != nil {
		return err
	}

	// enumerate packs user has created
	if len(userPacks.IsCreator) > 0 {
		text += "Here are the gif packs you have created:\n"

		for i, pack := range userPacks.IsCreator {
			text += fmt.Sprintf("%d. %s\n", i+1, pack.Name)
		}
	} else {
		text += "You have not created any gif packs.\n"
	}

	// enumerate packs user is a contributor to
	if len(userPacks.IsContributor) > 0 {
		text += "Here are the gif packs you are a contributor to:\n"

		for i, pack := range userPacks.IsContributor {
			text += fmt.Sprintf("%d. %s\n", i+1, pack.Name)
		}
	} else {
		text += "You are not a contributor to any gif packs."
	}

	reply := tgbotapi.NewMessage(chatID, text)
	_, err = bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func cmdSubscribeHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	userID := message.From.ID
	chatID := message.Chat.ID

	var text string
	done := false
	if packName := message.CommandArguments(); packName != "" {
		subscribed, err := Subscribe(ctx, packName, userID)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
				done = true
			} else {
				return err
			}
		} else {
			if subscribed {
				text = "Great! You have been subscribed to this gif pack!"
				done = true
			} else {
				text = "Don't worry, you are already subscribed to this gif pack!"
				done = true
			}
		}
	} else {
		text = "What is the name of the gif pack you want to subscribe to?"
	}

	if !done {
		state := ConversationState{
			State: stateSubscribeWaitPackName,
		}

		err := SetConversationState(ctx, chatID, userID, state)
		if err != nil {
			return err
		}
	}

	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if !done {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func cmdUnsubscribeHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	userID := message.From.ID
	chatID := message.Chat.ID

	var text string
	done := false
	if packName := message.CommandArguments(); packName != "" {
		unsubscribed, err := Unsubscribe(ctx, packName, userID)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. Pack names can only contain letter, numbers, hyphens and underscores."
				done = true
			} else {
				return err
			}
		} else {
			if unsubscribed {
				text = "Great! You have been unsubscribed from that gif pack."
				done = true
			} else {
				text = "Don't worry, it seems like you were never subscribed to that gif pack in the first place."
				done = true
			}
		}
	} else {
		text = "What is the name of the gif pack you want to unsubscribe from?"
	}

	if !done {
		state := ConversationState{
			State: stateUnsubscribeWaitPackName,
		}

		err := SetConversationState(ctx, chatID, userID, state)
		if err != nil {
			return err
		}
	}

	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID

		if !done {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func cmdSubscriptionsHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	userID := message.From.ID
	chatID := message.Chat.ID

	var text string
	subscriptions, err := MySubscriptions(ctx, userID)
	if err != nil {
		return err
	}

	if len(subscriptions) > 0 {
		text = "Here are the packs you are currently subscribed to: \n"

		for i, subscription := range subscriptions {
			text += fmt.Sprintf("%d. %s\n", i+1, subscription.Pack)
		}
	} else {
		text = "Oops! It looks like you haven't subscribed to any packs yet."
	}

	reply := tgbotapi.NewMessage(chatID, text)
	_, err = bot.Send(reply)
	if err != nil {
		return err
	}

	return err
}

func cmdNewGifHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	userID := message.From.ID
	chatID := message.Chat.ID

	var reply tgbotapi.MessageConfig
	packName := message.CommandArguments()

	if packName != "" {
		pack, err := GetPack(ctx, packName)
		if err != nil {
			if err == ErrNotFound {
				text := "Oh no! That pack does not exist. Did you spell it correctly?"
				reply = tgbotapi.NewMessage(chatID, text)
			} else {
				return err
			}
		} else {
			if pack.Creator == userID {
				state := ConversationState{
					State: stateNewGifWaitGif,
					Data: map[string]string{
						"packName": packName,
					},
				}

				err := SetConversationState(ctx, chatID, userID, state)
				if err != nil {
					return err
				}

				text := "Please send me the gif you want to add to this pack."
				reply = tgbotapi.NewMessage(chatID, text)

			} else {
				text := "Oops, it seems like you are not the creator of this pack. Only the pack creator can add gifs to a pack."
				reply = tgbotapi.NewMessage(chatID, text)
			}
		}

	} else {
		state := ConversationState{
			State: stateNewGifWaitPackName,
		}

		err := SetConversationState(ctx, chatID, userID, state)
		if err != nil {
			return err
		}

		text := "Which pack do you want to add a new gif to?"
		reply = tgbotapi.NewMessage(chatID, text)

	}

	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
		reply.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  true,
		}
	}

	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func cmdDeleteGifHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	userID := message.From.ID
	chatID := message.Chat.ID

	var state ConversationState
	var text string
	done := false
	if packName := message.CommandArguments(); packName != "" {
		_, err := GetPack(ctx, packName)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
				done = true
			} else {
				return err
			}
		} else {
			state = ConversationState{
				State: stateDeleteGifWaitGif,
				Data: map[string]string{
					"packName": packName,
				},
			}

			text = "Please send me the gif you want to delete from this pack."
		}
	} else {
		state = ConversationState{
			State: stateDeleteGifWaitPackName,
		}

		text = "Which gif pack do you want to delete a gif from?"
	}

	if !done {
		err := SetConversationState(ctx, chatID, userID, state)
		if err != nil {
			return err
		}
	}

	reply := tgbotapi.NewMessage(chatID, text)

	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
		if !done {
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}
	}

	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func cmdVersionHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("Saved GIFs Bot version %s", Version))
	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func cmdCancelHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	reply := tgbotapi.NewMessage(chatID, "Command cancelled!")
	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}
