# piktagbot

A Telegram bot that helps users with sticker management. 
You can attach tags to stickers and then search your stickers by tags in inline mode.

Bot settings are imported from env vars:
* `APP_ID` — `app_id` from [my.telegram.org](https://my.telegram.org/apps)
* `APP_HASH` — `app_hash` from [my.telegram.org](https://my.telegram.org/apps)
* `BOT_TOKEN` — bot token from [BotFather](https://t.me/BotFather)
* `SESSION_FILE` — path to session .json file (existing or not)
* `dbUser` — database user
* `dbPass` — database password