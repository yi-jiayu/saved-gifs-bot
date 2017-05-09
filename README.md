[![Build Status](https://travis-ci.org/yi-jiayu/saved-gifs-bot.svg?branch=master)](https://travis-ci.org/yi-jiayu/saved-gifs-bot)
[![codecov](https://codecov.io/gh/yi-jiayu/saved-gifs-bot/branch/master/graph/badge.svg)](https://codecov.io/gh/yi-jiayu/saved-gifs-bot)

# Saved GIFs Bot
A bot for organising your GIFs on Telegram

## Motivation
Do you find yourself scrolling through your entire Telegram GIF drawer looking for a particular GIF you want to send? 
Do you wonder why GIFs can't be categorised and sorted like stickers? Saved GIFs Bot is here to help! 

## Features
- Create GIF packs just like you would create sticker packs
- Filter GIFs by packs just like stickers
- Tag and search GIFs with keywords
 
## Getting started
1. Add [@SavedGIFsBot](https://t.me/SavedGIFsBot) on Telegram
2. Use `/newpack` to create a new GIF pack
3. Use `/newgif` to add GIFs to your GIF pack
4. Use `/subscribe` to subscribe to GIF packs
5. Search your GIFs in any chat by sending inline queries to Saved GIFs Bot.

## Query format
Saved GIFs Bot understands inline queries of the form `@SavedGIFsBot [pack-name] [keywords]`".
- If a GIF pack named `pack-name` exists, Saved GIFs Bot will show GIFs from that pack.
- If `pack-name` is `-`, Saved GIFs Bot will show GIFs from all packs you are subscribed to.
- If `keywords` are provided, only GIFs which were tagged with `keywords` will be shown.
- An empty query will show GIFs from all the packs you are subscribed to.

## Notes
- Hosted on Google App Engine Go Standard Environment
- Saved GIFs Bot is still in active development and may be unstable.

## [Change log](CHANGELOG.md)
