package saved_gifs_bot

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
	"google.golang.org/appengine/search"
	"google.golang.org/appengine/urlfetch"
)

type InlineQueryResultCachedMpeg4Gif struct {
	Type                string                         `json:"type"`
	Id                  string                         `json:"id"`
	Mpeg4FileId         string                         `json:"mpeg4_file_id"`
	Title               string                         `json:"title"`
	Caption             string                         `json:"caption"`
	ReplyMarkup         *tgbotapi.InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}                    `json:"input_message_content,omitempty"`
}

func NewInlineQueryResultCachedMpeg4Gif(id, fileId string) InlineQueryResultCachedMpeg4Gif {
	return InlineQueryResultCachedMpeg4Gif{
		Type:        "mpeg4_gif",
		Id:          id,
		Mpeg4FileId: fileId,
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
		if command := message.Command(); command != "" {
			if handler, exists := commandHandlers[command]; exists {
				err := handler(ctx, &bot, message)
				if err != nil {
					SomethingWentWrong(ctx, &bot, message, err)
				}
			}
		} else if document := message.Document; document != nil {
			if handler, exists := documentHandlers[document.MimeType]; exists {
				err := handler(ctx, &bot, message)
				if err != nil {
					SomethingWentWrong(ctx, &bot, message, err)
				}
			}
		} else if message.Text != "" {
			var textHandler MessageHandler = func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
				chatId := message.Chat.ID
				userId := message.From.ID
				text := message.Text

				state, err := GetConversationState(ctx, chatId, userId)
				if err != nil {
					return err
				}

				// continuation of newgif
				if state["action"] == "newgif" && state["pack"] != "" && state["fileId"] != "" {
					pack := state["pack"]
					gif := Gif{
						FileID:   search.Atom(state["fileId"]),
						Keywords: text,
					}

					err := NewGif(ctx, pack, userId, gif)
					if err != nil {
						return err
					}

					text := "Great! A new gif has been added to your gif pack."
					reply := tgbotapi.NewMessage(chatId, text)
					_, err = bot.Send(reply)
					if err != nil {
						return err
					}
				}

				return nil
			}

			err := textHandler(ctx, &bot, message)
			if err != nil {
				log.Errorf(ctx, "%v", err)
			}
		}
	} else if inlineQuery := update.InlineQuery; inlineQuery != nil {
		id := inlineQuery.ID

		config := tgbotapi.InlineConfig{
			InlineQueryID: id,
			Results: []interface{}{
				NewInlineQueryResultCachedMpeg4Gif("CgADBQADAgAD_AfhV7d7t0mZPmy8Ag", "CgADBQADAgAD_AfhV7d7t0mZPmy8Ag"),
				NewInlineQueryResultCachedMpeg4Gif("CgADBQADLgADn0JAV_DttqWfPpaKAg", "CgADBQADLgADn0JAV_DttqWfPpaKAg"),
				NewInlineQueryResultCachedMpeg4Gif("CgADBQADAQADyXiIVajLNBrhWeT8Ag", "CgADBQADAQADyXiIVajLNBrhWeT8Ag"),
				NewInlineQueryResultCachedMpeg4Gif("CgADBAADhEAAAlsdZAfSEgVmHiiiFAI", "CgADBAADhEAAAlsdZAfSEgVmHiiiFAI"),
			},
		}

		resp, err := bot.AnswerInlineQuery(config)
		if err != nil {
			log.Errorf(ctx, "%v", resp)
			log.Errorf(ctx, "%v", err)
		}
	}
}

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
