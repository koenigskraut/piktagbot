# piktagbot

A Telegram bot that helps users with sticker management. 
You can attach tags to stickers and then search your stickers by tags in inline mode.

Furthermore, now you can sort recent or tagged stickers in any order you choose with simple WebApp interface.

There are two executables in `cmd`. Settings are env vars, for both executables they are:
* `dbUser` — database user
* `dbPass` — database password
* `APP_HASH` — `app_id` from [my.telegram.org](https://my.telegram.org/apps)
* `APP_ID` — `app_hash` from [my.telegram.org](https://my.telegram.org/apps)
* `BOT_TOKEN` — bot token from [BotFather](https://t.me/BotFather)
* `SESSION_FILE` — path to session .json file (existing or not)
* `DOMAIN` — web app domain
* `APP_PORT` —  web app port

and only for web app there are: 
* `CERT_FILE` — TLS certificate file
* `KEY_FILE` — TLS private key file
* `STICKER_PATH` — folder where to store caches stickers
