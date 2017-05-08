package saved_gifs_bot

import (
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/search"
)

var Transduce MessageHandler = func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatId := message.Chat.ID
	userId := message.From.ID

	state, err := GetConversationState(ctx, chatId, userId)
	if err != nil {
		return err
	}

	switch {
	// newgif waiting for pack name
	case state["action"] == "newgif" && state["pack"] == "":
		user := message.From.ID
		chatId := message.Chat.ID

		var reply tgbotapi.MessageConfig
		if packName := message.Text; packName != "" {

			pack, err := GetPack(ctx, packName)
			if err != nil {
				if err == ErrNotFound {
					text := "Oh no! That pack does not exist. Did you spell it correctly?"
					reply = tgbotapi.NewMessage(chatId, text)
				} else {
					return err
				}
			} else {
				if pack.Creator == user {
					state := map[string]string{
						"action": "newgif",
						"pack":   packName,
					}
					err := SetConversationState(ctx, chatId, user, state)
					if err != nil {
						return err
					} else {
						text := "Please send me the gif you want to add to this pack."
						reply = tgbotapi.NewMessage(chatId, text)
					}
				} else {
					text := "Oops, it seems like you are not the creator of this pack. Only the pack creator can add gifs to a pack."
					reply = tgbotapi.NewMessage(chatId, text)
				}
			}
		} else {
			text := "Oops! I was waiting for you to send me the name of the gif pack you want to add a new gif to."
			reply = tgbotapi.NewMessage(chatId, text)
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

		// newgif waiting for gif
	case state["action"] == "newgif" && state["pack"] != "" && state["fileId"] == "":
		var reply tgbotapi.MessageConfig
		if document := message.Document; document != nil {
			state["fileId"] = document.FileID

			err = SetConversationState(ctx, chatId, userId, state)
			if err != nil {
				return err
			}

			text := "Alright, now send me some keywords that describe this gif."
			reply = tgbotapi.NewMessage(chatId, text)
		} else {
			text := "Oops, I was waiting for you to send me a gif."
			reply = tgbotapi.NewMessage(chatId, text)
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

		// newgif waiting for keywords
	case state["action"] == "newgif" && state["pack"] != "" && state["fileId"] != "":
		var reply tgbotapi.MessageConfig
		done := false
		if text := message.Text; text != "" {
			pack := state["pack"]
			gif := Gif{
				Pack:     search.Atom(state["pack"]),
				FileID:   search.Atom(state["fileId"]),
				Keywords: text,
			}

			err := NewGif(ctx, pack, userId, gif)
			if err != nil {
				return err
			}

			t := "Great! A new gif has been added to your gif pack."
			reply = tgbotapi.NewMessage(chatId, t)
			done = true
		} else {
			text := "Oops, I was waiting for you to send me some keywords for this gif."
			reply = tgbotapi.NewMessage(chatId, text)
		}

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

		if done {
			// clear state now that we are done with this flow
			state = nil
			err = SetConversationState(ctx, chatId, userId, state)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}
