package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type InlineQueryResultCachedMpeg4Gif struct {
	Type                string                         `json:"type"`          // required
	ID                  string                         `json:"id"`            // required
	Mpeg4FileID         string                         `json:"mpeg4_file_id"` // required
	Title               string                         `json:"title"`
	Caption             string                         `json:"caption"`
	ReplyMarkup         *tgbotapi.InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}                    `json:"input_message_content,omitempty"`
}

func NewInlineQueryResultCachedMpeg4Gif(id, fileID string) InlineQueryResultCachedMpeg4Gif {
	return InlineQueryResultCachedMpeg4Gif{
		Type:        "mpeg4_gif",
		ID:          id,
		Mpeg4FileID: fileID,
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf(ctx, "%v", err)

		// for now we will just return a 200 status to all webhooks so that telegram does not redeliver them
		// w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// log update
	log.Infof(ctx, string(bytes))

	var update tgbotapi.Update
	err = json.Unmarshal(bytes, &update)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		return
	}

	client := urlfetch.Client(ctx)
	bot := tgbotapi.BotAPI{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Client: client,
	}

	if message := update.Message; message != nil {
		// handle a new command
		if command := message.Command(); command != "" {
			if handler, exists := commandHandlers[command]; exists {
				// clear bot state before a new command
				err := ClearConversationState(ctx, message.Chat.ID, message.From.ID)
				if err != nil {
					SomethingWentWrong(ctx, &bot, message, err)
					return
				}

				err = handler(ctx, &bot, message)
				if err != nil {
					SomethingWentWrong(ctx, &bot, message, err)
					return
				}
			}
		} else {
			// or continue based on the previous conversation state
			err := Transduce(ctx, &bot, message)
			if err != nil {
				SomethingWentWrong(ctx, &bot, message, err)
				return
			}
		}

		return
	}

	if inlineQuery := update.InlineQuery; inlineQuery != nil {
		HandleInlineQuery(ctx, &bot, inlineQuery)
		return
	}
}

// SomethingWentWrong replies to a message saying that something went wrong and provides a request id for reporting the
// error.
func SomethingWentWrong(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, err error) {
	log.Errorf(ctx, "%v", err)
	text := fmt.Sprintf("Oh no! Something went wrong. Request Id: `%s`", appengine.RequestID(ctx))
	reply := tgbotapi.NewMessage(message.Chat.ID, text)
	reply.ParseMode = "markdown"
	_, err2 := bot.Send(reply)
	if err2 != nil {
		log.Errorf(ctx, "%v", err2)
	}
}

func init() {
	http.HandleFunc("/", rootHandler)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		http.HandleFunc("/"+token, webhookHandler)
	}
}
