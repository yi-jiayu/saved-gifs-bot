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
	// todo: newgif waiting for pack name

	// newgif waiting for gif
	case state["action"] == "newgif" && state["pack"] != "" && state["fileId"] == "":
		if document := message.Document; document != nil {
			state["fileId"] = document.FileID

			err = SetConversationState(ctx, chatId, userId, state)
			if err != nil {
				return err
			}

			text := "Alright, now send me some keywords that describe this gif."
			reply := tgbotapi.NewMessage(chatId, text)
			_, err := bot.Send(reply)
			if err != nil {
				return err
			}

			return nil
		} else {
			text := "Oops, I was waiting for you to send me a gif."
			reply := tgbotapi.NewMessage(chatId, text)
			_, err := bot.Send(reply)
			if err != nil {
				return err
			}

			return nil
		}
		// newgif waiting for keywords
	case state["action"] == "newgif" && state["pack"] != "" && state["fileId"] != "":
		if text := message.Text; text != "" {
			pack := state["pack"]
			gif := Gif{
				FileID:   search.Atom(state["fileId"]),
				Keywords: text,
			}

			err := NewGif(ctx, pack, userId, gif)
			if err != nil {
				return err
			}

			t := "Great! A new gif has been added to your gif pack."
			reply := tgbotapi.NewMessage(chatId, t)
			_, err = bot.Send(reply)
			if err != nil {
				return err
			}

			// clear state now that we are done with this flow
			state = nil
			err = SetConversationState(ctx, chatId, userId, state)
			if err != nil {
				return err
			}
		} else {
			text := "Oops, I was waiting for you to send me some keywords for this gif."
			reply := tgbotapi.NewMessage(chatId, text)
			_, err := bot.Send(reply)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}
