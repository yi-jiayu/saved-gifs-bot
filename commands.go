package saved_gifs_bot

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type messageHandler func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message)

var commands = map[string]messageHandler{
	"newpack": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := fmt.Sprintf("%d", message.From.ID)
		chatId := message.Chat.ID
		name := message.CommandArguments()

		var text string
		created, err := NewList(ctx, name, user)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid list name. A list name can only contain letters, numbers, hyphens and underscores."
			} else {
				log.Errorf(ctx, "%v", err)
				text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
			}
		} else {
			if created {
				text = "Great! Your gif list has been created."
			} else {
				text = "Oh no! That list name has already been taken. Can you think of another one?"
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"mypacks": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := fmt.Sprintf("%d", message.From.ID)
		chatId := message.Chat.ID

		var text string
		lists, err := MyLists(ctx, user)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
		} else {
			if len(lists) > 0 {
				text = "Here are the lists you have created: \n"

				for i, list := range lists {
					text += fmt.Sprintf("%d. %s\n", i+1, list.Name)
				}
			} else {
				text = "Oops! It looks like you haven't created any lists yet."
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"subscribe": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := fmt.Sprintf("%d", message.From.ID)
		chatId := message.Chat.ID

		var text string
		list := message.CommandArguments()
		subscribed, err := Subscribe(ctx, list, user)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any list with that name."
			} else {
				log.Errorf(ctx, "%v", err)
				text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
			}
		} else {
			if subscribed {
				text = "Great! You have been subscribed to this list!"
			} else {
				text = "Don't worry, you are already subscribed to this list!"
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"unsubscribe": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := fmt.Sprintf("%d", message.From.ID)
		chatId := message.Chat.ID

		var text string
		list := message.CommandArguments()
		unsubscribed, err := Unsubscribe(ctx, list, user)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid list name. List names can only contain letter, numbers, hyphens and underscores."
			} else {
				log.Errorf(ctx, "%v", err)
				text = fmt.Sprintf("Oh no! Something went wrong. Request Id: %s", appengine.RequestID(ctx))
			}
		} else {
			if unsubscribed {
				text = "Great! You have been unsubscribed from that list."
			} else {
				text = "Don't worry, it seems like you were never subscribed to that list in the first place."
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%v", err)
		}
	},
	"subscriptions": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
		user := fmt.Sprintf("%d", message.From.ID)
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
					text += fmt.Sprintf("%d. %s\n", i+1, subscription.List)
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
