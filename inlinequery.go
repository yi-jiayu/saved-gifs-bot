package saved_gifs_bot

import (
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func HandleInlineQuery(ctx context.Context, bot *tgbotapi.BotAPI, inlineQuery *tgbotapi.InlineQuery) {
	id := inlineQuery.ID
	userId := inlineQuery.From.ID
	query := inlineQuery.Query

	if query == "" {
		return
	}

	gifs, err := SearchGifs(ctx, userId, query)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	if len(gifs) == 0 {
		return
	}

	// deduplicate results
	gifsMap := make(map[string]int)
	for _, gif := range gifs {
		gifsMap[string(gif.FileID)] = 0
	}

	var results []interface{}
	for fileId := range gifsMap {
		id := string(fileId)
		results = append(results, NewInlineQueryResultCachedMpeg4Gif(id, id))
	}

	config := tgbotapi.InlineConfig{
		InlineQueryID: id,
		Results:       results,
	}

	resp, err := bot.AnswerInlineQuery(config)
	if err != nil {
		log.Errorf(ctx, "%v", resp)
		log.Errorf(ctx, "%v", err)
	}
}
