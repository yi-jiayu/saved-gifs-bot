package saved_gifs_bot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

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
	}
}

func init() {
	http.HandleFunc("/", rootHandler)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		http.HandleFunc("/"+token, webhookHandler)
	}
}
