# Run your own server

## Deploy on your own server

Install [Go](https://go.dev/doc/install) on your host machine.  

Initialize server with folders and systemd service. Tested on Debian-based systems:
```bash
$ make init_server host=user@example.com salt=$(head -c 32 /dev/urandom | base64)
```

Deploy a systemd service:
```bash
$ make deploy_systemd host=<YOUR_SSH_HOST>
```

That's all :)  

## Run your own Telegram Bot
1) Install [Go](https://go.dev/doc/install)
2) Register new telegram bot via [@BotFather](https://t.me/BotFather)
3) Add `BOT_API_TOKEN=<YOUR_TELEGRAM_API_TOKEN>` line to `.env` file
4) Redeploy/relaunch the server

Bot's artifacts can be seen in `./storage/<USER_ID>` folder.  

## Run your own Feishu bot

The Feishu integration uses Feishu's long connection mode, so the server opens
an outbound WebSocket connection to Feishu. You do not need to expose a webhook
endpoint for incoming bot messages.

1) Create a custom Feishu app in Feishu Developer Console.
2) Enable bot capability for the app.
3) Subscribe to the `im.message.receive_v1` event.
   If `FEISHU_ENABLE_CARD_ACTIONS=true`, also subscribe to
   `card.action.trigger`.
4) Set the event/callback subscription mode to long connection.
5) Add the app to a Feishu chat and send the bot a direct message.
6) Add the following variables to `.env`:

```bash
FEISHU_APP_ID=<YOUR_FEISHU_APP_ID>
FEISHU_APP_SECRET=<YOUR_FEISHU_APP_SECRET>
FEISHU_ALLOWED_OPEN_IDS=<YOUR_OPEN_ID>
FEISHU_DEFAULT_USER_ID=10001
FEISHU_ENABLE_CARD_ACTIONS=false
STORAGE_DIR=./storage
```

`FEISHU_DEFAULT_USER_ID` is the numeric files.md user directory to write to.
For the example above, messages are saved under `./storage/10001`. If it is not
set, the server derives a stable numeric ID from the Feishu sender `open_id`.

For a personal vault where `STORAGE_DIR` should be the vault root itself, set:

```bash
SINGLE_USER_MODE=true
SINGLE_USER_ID=10001
CONFIG_FILENAME=.filesmd.json
STORAGE_DIR=/path/to/Obsidian
```

In this mode, new files are written directly under `/path/to/Obsidian` instead
of `/path/to/Obsidian/10001`.

For a personal setup, keep `FEISHU_ALLOWED_OPEN_IDS` set to your own `open_id`.
If it is empty, every Feishu sender that can message the bot is accepted.

The MVP saves regular text messages to `Chat.md`. The existing `jj` suffix still
saves a message to journal.

## Linking a new device
1) Open telegram bot
2) Open `/app`
3) Open the link in your browser
4) Device is now linked

### Additional bot's settings
1) For search functionality, enable `Inline Mode` for your bot in [@BotFather](https://t.me/BotFather)
2) Press "Edit Commands", and send the following list:
```
chat - 🏠 Home
files - 📄 Files
dirs - 🗂 Dirs
checklists - ☑️ Checklists
schedule - 📆 Schedule
postpone - 🦥 Postpone
rename - ✏️ Rename
move - ➡️ Move
app - 🔗 Open in app
settings - ⚙️ Settings
help - 📕 Help
```

## Hosting the bot on you local computer
You can host the bot locally, because it doesn't expose any ports to the outside world (if you don't use habits functionality).  
It communicates with Telegram using pull API.

Create a symlink to your local folder with `.md` files for convenience:  
`ln -s <YOUR_EXISTING_DIR_WITH_MD_FILES> storage/<USER_ID>`

## Transfer files to another server

1) Backup your data (`/app/storage`)
2) Be sure that all client app fully synced with the server (bring the app in the focus)
3) Stop bot on old server, so no new files would be created.
4) Compress all the files on one server: `tar -czvf storage.tar.gz storage`
5) `scp` the file to your host machine: `scp SSH_HOST:/app/storage.tar.gz .`
6) `scp` the file to your target machine

Synchronization is relying on `mtime`, so after compressing/decompressing the flag wouldn't be lost.

1) `cd /opt/files.md`
2) `tar -czvf tokens.tar.gz tokens`
3) `scp` to same dir on target machine

We don't need to transfer fslog (renames), if we're certain that all clients read the log.

1) Extract all files on new server
2) Transfer `BOT_API_TOKEN`
3) Launch server
4) Execute `localStorage.setItem('ApiHost', 'YOUR_NEW_API_HOST');` in your PWA applications
5) Make sure that all files are available
6) Cleanup the oldserver

## Maintenance notes
Add this to your crontab (`crontab -e`) for daily git backups:
`0 0 * * * cd /app/storage/<YOUR_TELEGRAM_ID> && git add . && git commit -m "$(date +\%d.\%m.\%Y)"`

Execute `git init` in your folder before that, to init a git repository.

If you have non-ASCI character in filenames, disable quoting:
`git config --global core.quotePath false`

Systemd journal:  
`sudo journalctl -u filesmd`

Find forbidden character in filenames (can be executed in user's storage folder):
`find . -name '*[<>:"|\?*]*'`

Remove forbidden filename characters:
```bash
find . -type f -name '*[<>:"|\?*]*' -print0 | while IFS= read -r -d '' f; do
  dir=$(dirname "$f")
  base=$(basename "$f")
  newbase="${base//[<>:\"|\\?*]/}"
  [ "$base" != "$newbase" ] && [ -n "$newbase" ] && mv -n -- "$f" "$dir/$newbase"
done
```
