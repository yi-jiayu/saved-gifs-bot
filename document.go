package saved_gifs_bot

import (
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
)

var documentHandlers = map[string]MessageHandler{
	"video/mp4": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
		chatId := message.Chat.ID
		userId := message.From.ID
		document := message.Document

		state, err := GetConversationState(ctx, chatId, userId)
		if err != nil {
			return err
		}

		// continuation from newgif
		if state["action"] == "newgif" && state["pack"] != "" {
			state["fileId"] = document.FileID

			err = SetConversationState(ctx, chatId, userId, state)
			if err != nil {
				return err
			}

			text := "Alright, now send me some keywords that describe this gif."
			reply := tgbotapi.NewMessage(chatId, text)
			bot.Send(reply)
			return nil
		}

		return nil
	},
}
