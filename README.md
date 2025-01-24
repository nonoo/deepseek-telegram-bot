# deepseek-telegram-bot

This bot lets you use the [DeepSeek Chat](https://chat.deepseek.com/) through
its [API](https://api-docs.deepseek.com/api/deepseek-api/).

An `API key` and a topped up balance are needed to use the bot. You can get an
API key by registering on the [DeepSeek Platform](https://platform.deepseek.com/).

Tested on Linux, but should be able to run on other operating systems.

## Compiling

You'll need Go installed on your computer. Install a recent package of `golang`.
Then:

```
go get github.com/nonoo/deepseek-telegram-bot
go install github.com/nonoo/deepseek-telegram-bot
```

This will typically install `deepseek-telegram-bot` into `$HOME/go/bin`.

Or just enter `go build` in the cloned Git source repo directory.

## Prerequisites

Create a Telegram bot using [BotFather](https://t.me/BotFather) and get the
bot's `token`.

## Running

You can get the available command line arguments with `-h`.
Mandatory arguments are:

- `-ds-api-key`: set this to your DeepSeek `API key`
- `-bot-token`: set this to your Telegram bot's `token`
- `-chat-cmd`: set this to the command to use to send messages to the bot
- `-ds-initial-prompt`: set this to the initial prompt to send to the DeepSeek
  API
- `-ds-temperature`: set this to the [temperature](https://api-docs.deepseek.com/quick_start/parameter_settings)
  to send to the DeepSeek API

Set your Telegram user ID as an admin with the `-admin-user-ids` argument.
Admins will get a message when the bot starts.

Other user/group IDs can be set with the `-allowed-user-ids` and
`-allowed-group-ids` arguments. IDs should be separated by commas.

You can get Telegram user IDs by writing a message to the bot and checking
the app's log, as it logs all incoming messages.

All command line arguments can be set through OS environment variables.
Note that using a command line argument overwrites a setting by the environment
variable. Available OS environment variables are:

- `DS_API_KEY`
- `BOT_TOKEN`
- `CHAT_CMD`
- `DS_INITIAL_PROMPT`
- `DS_TEMPERATURE`
- `DS_HISTORY_SIZE`
- `ALLOWED_USERIDS`
- `ADMIN_USERIDS`
- `ALLOWED_GROUPIDS`

## Supported commands

- `chat (msg)` - send a message to the DeepSeek API
- `dsbalance` - query the current API account balance
- `dshelp` - show the help

## Contributors

- Norbert Varga [nonoo@nonoo.hu](mailto:nonoo@nonoo.hu)

## Donations

If you find this bot useful then [buy me a beer](https://paypal.me/ha2non). :)
