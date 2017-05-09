package main

import (
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// HandleInlineQuery handles incoming inline queries.
func HandleInlineQuery(ctx context.Context, bot *tgbotapi.BotAPI, inlineQuery *tgbotapi.InlineQuery) {
	inlineQueryID := inlineQuery.ID
	userID := inlineQuery.From.ID
	query := inlineQuery.Query

	gifs, err := SearchGifs(ctx, userID, query)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	results := make([]interface{}, 0)
	if len(gifs) > 0 {
		// deduplicate results
		gifsMap := make(map[string]int)
		for _, gif := range gifs {
			gifsMap[string(gif.FileID)] = 0
		}

		for fileID := range gifsMap {
			id := string(fileID)
			results = append(results, NewInlineQueryResultCachedMpeg4Gif(id, id))
		}
	}

	config := tgbotapi.InlineConfig{
		InlineQueryID: inlineQueryID,
		Results:       results,
		IsPersonal:    true,
	}

	resp, err := bot.AnswerInlineQuery(config)
	if err != nil {
		log.Errorf(ctx, "%v", resp)
		log.Errorf(ctx, "%v", err)
	}
}
