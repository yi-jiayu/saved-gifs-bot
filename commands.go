package saved_gifs_bot

import (
	"fmt"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
)

var commandHandlers = map[string]MessageHandler{
	"newpack": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
		user := message.From.ID
		chatId := message.Chat.ID
		name := message.CommandArguments()

		var text string
		created, err := NewPack(ctx, name, user)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. A pack name can only contain letters, numbers, hyphens and underscores."
			} else {
				return err
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
			return err
		}

		return nil
	},
	"mypacks": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		packs, err := MyPacks(ctx, user)
		if err != nil {
			return err
		} else {
			if len(packs) > 0 {
				text = "Here are the gif packs you have created: \n"

				for i, pack := range packs {
					text += fmt.Sprintf("%d. %s\n", i+1, pack.Name)
				}
			} else {
				text = "Oops! It looks like you haven't created any gif packs yet."
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			return err
		}

		return nil
	},
	"subscribe": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		pack := message.CommandArguments()
		subscribed, err := Subscribe(ctx, pack, user)
		if err != nil {
			if err == ErrNotFound {
				text = "Oops! There doesn't seem to be any gif pack with that name."
			} else {
				return err
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
			return err
		}

		return nil
	},
	"unsubscribe": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		pack := message.CommandArguments()
		unsubscribed, err := Unsubscribe(ctx, pack, user)
		if err != nil {
			if err == ErrInvalidName {
				text = "Oh no! That was not a valid pack name. Pack names can only contain letter, numbers, hyphens and underscores."
			} else {
				return err
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
			return err
		}

		return nil
	},
	"subscriptions": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
		user := message.From.ID
		chatId := message.Chat.ID

		var text string
		subscriptions, err := MySubscriptions(ctx, user)
		if err != nil {
			return err
		} else {
			if len(subscriptions) > 0 {
				text = "Here are the packs you are currently subscribed to: \n"

				for i, subscription := range subscriptions {
					text += fmt.Sprintf("%d. %s\n", i+1, subscription.Pack)
				}
			} else {
				text = "Oops! It looks like you haven't subscribed to any packs yet."
			}
		}

		reply := tgbotapi.NewMessage(chatId, text)
		_, err = bot.Send(reply)
		if err != nil {
			return err
		}

		return err
	},
	"newgif": func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
		user := message.From.ID
		chatId := message.Chat.ID

		var reply tgbotapi.MessageConfig
		packName := message.CommandArguments()

		if packName != "" {
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
			state := map[string]string{
				"action": "newgif",
			}
			err := SetConversationState(ctx, chatId, user, state)
			if err != nil {
				return err
			} else {
				text := "Which pack do you want to add a new gif to?"
				reply = tgbotapi.NewMessage(chatId, text)
			}
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
	},
}

type MessageHandler func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error
