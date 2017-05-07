package saved_gifs_bot

import (
	"fmt"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type messageHandler func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message)

var commands = map[string]messageHandler{
	"newpack": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := message.From.ID
		chatId := message.Chat.ID
		name := message.CommandArguments()

		var text string
		created, err := NewPack(ctx, name, user)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
			} else {
				log.Errorf(ctx, "%v", err)
				text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
			}
		} else {
			if created {
				text = "Great! Your gif pack has been created."
			} else {
				text = "Oh no! That pack name has already been taken. Can you think of another one?"
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"mypacks": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		packs, err := MyPacks(ctx, user)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
		} else {
			if len(packs) > 0 {
				text = "Here are the gif packs you have created: \n"

				for i, list := range packs {
					text += fmt.Sprintf("%d. %s\n", i+1, list.Name)
				}
			} else {
				text = "Oops! It looks like you haven't created any gif packs yet."
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"subscribe": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		pack := message.CommandArguments()
		subscribed, err := Subscribe(ctx, pack, user)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
			} else {
				log.Errorf(ctx, "%v", err)
				text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
			}
		} else {
			if subscribed {
				text = "Great! You have been subscribed to this gif pack!"
			} else {
				text = "Don't worry, you are already subscribed to this gif pack!"
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"unsubscribe": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		pack := message.CommandArguments()
		unsubscribed, err := Unsubscribe(ctx, pack, user)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. Pack names can only contain letter, numbers, hyphens and underscores."
			} else {
				log.Errorf(ctx, "%v", err)
				text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
			}
		} else {
			if unsubscribed {
				text = "Great! You have been unsubscribed from that gif pack."
			} else {
				text = "Don't worry, it seems like you were never subscribed to that gif pack in the first place."
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"subscriptions": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		subscriptions, err := MySubscriptions(ctx, user)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
		} else {
			if len(subscriptions) > 0 {
				text = "Here are the lists you are currently subscribed to: \n"

				for i, subscription := range subscriptions {
					text += fmt.Sprintf("%d. %s\n", i+1, subscription.Pack)
				}
			} else {
				text = "Oops! It looks like you haven't subscribed to any lists yet."
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
}
