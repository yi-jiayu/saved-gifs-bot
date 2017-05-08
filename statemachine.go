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

	done := false

	switch {
	case state["action"] == "newgif" && state["pack"] == "": // newgif waiting for pack name
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
	case state["action"] == "newgif" && state["pack"] != "" && state["fileId"] == "": // newgif waiting for gif
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
	case state["action"] == "newgif" && state["pack"] != "" && state["fileId"] != "": // newgif waiting for keywords
		var reply tgbotapi.MessageConfig
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

		return nil
	case state["action"] == "newpack": // newpack waiting for pack name
		var text string
		if packName := message.Text; packName != "" {
			created, err := NewPack(ctx, packName, userId)
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
			text = "Oops! I was waiting for you to send me a name for your new gif pack."
		}

		reply := tgbotapi.NewMessage(chatId, text)
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
	case state["action"] == "subscribe": // subscribe waiting for pack name
		var text string
		if packName := message.Text; packName != "" {
			subscribed, err := Subscribe(ctx, packName, userId)
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
			text = "Oops, I was waiting for you to send me the name of the gif pack you want to subcribe to."
		}

		reply := tgbotapi.NewMessage(chatId, text)
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
	case state["action"] == "unsubscribe": // unsubscribe waiting for pack name
		var text string
		if packName := message.Text; packName != "" {
			unsubscribed, err := Unsubscribe(ctx, packName, userId)
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
			text = "Oops, I was waiting for you to send me the name of the gif pack you want to unsubscribe from."
		}

		reply := tgbotapi.NewMessage(chatId, text)
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
	case state["action"] == "deletegif" && state["pack"] == "": // deletegif waiting for pack name
		var text string
		if packName := message.Text; packName != "" {
			_, err := GetPack(ctx, packName)
			if err != nil {
				if err == ErrNotFound {
					text = "Oops! There doesn't seem to be any gif pack with that name."
					done = true
				} else {
					return err
				}
			} else {
				state["pack"] = packName
				err := SetConversationState(ctx, chatId, userId, state)
				if err != nil {
					return err
				}

				text = "Please send me the gif you want to delete from this pack."
			}
		} else {
			text = "Oops, I was waiting for you to send me the name of the gif pack you want to delete a gif from."
		}

		reply := tgbotapi.NewMessage(chatId, text)
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
	case state["action"] == "deletegif" && state["pack"] != "": // deletegif waiting for gif
		var text string
		if document := message.Document; document.MimeType == "video/mp4" {
			fileId := document.FileID

			deleted, err := DeleteGif(ctx, state["pack"], message.From.ID, fileId)
			if err != nil {
				if err == ErrNotFound {
					text = "Oops, there doesn't seem to be any gif pack with that name."
					done = true
				} else {
					return err
				}
			} else {
				if deleted > 0 {
					text = "Great, that gif has been deleted!"
					done = true
				} else {
					text = "Oops, I couldn't find that gif in this pack."
					done = true
				}
			}
		} else {
			text = "Oops, I was waiting for you to send me a gif to be deleted from this pack."
		}

		reply := tgbotapi.NewMessage(chatId, text)
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
	}

	if done {
		// reset state since we are done
		state = nil
		err = SetConversationState(ctx, chatId, userId, state)
		if err != nil {
			return err
		}
	}

	return nil
}
