package saved_gifs_bot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
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
			if handler, exists := commands[command]; exists {
				handler(ctx, &bot, message)
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

func init() {
	http.HandleFunc("/", rootHandler)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		http.HandleFunc("/"+token, webhookHandler)
	}
}
