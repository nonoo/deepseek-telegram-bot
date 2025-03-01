package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/exp/slices"
)

const errorStr = "❌ Error"

var dsClient *deepseek.Client
var telegramBot *bot.Bot
var cmdHandler cmdHandlerType

func convertLatexToHTML(latex string) string {
	latex = strings.ReplaceAll(latex, `\[`, `<blockquite>`)
	latex = strings.ReplaceAll(latex, `\]`, `</blockquote>`)
	latex = strings.ReplaceAll(latex, `\(`, `<blockquote>`)
	latex = strings.ReplaceAll(latex, `\)`, `</blockquote>`)

	return latex
}

func removeNewlineAfterBlockquote(s string) string {
	return strings.ReplaceAll(s, "<blockquote>\n", "<blockquote>")
}

func filterText(s string) string {
	return removeNewlineAfterBlockquote(filterHTML(convertLatexToHTML(s)))
}

func sendMessage(ctx context.Context, chatID int64, s string) (msg *models.Message, err error) {
	msg, err = telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      filterText(s),
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		msg, err = telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   filterText(s),
		})
		if err != nil {
			fmt.Println("  send error:", err)
			msg = nil
		}
	}
	return
}

func sendReplyToMessage(ctx context.Context, replyToMsg *models.Message, s string) (msg *models.Message, err error) {
	msg, err = telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ReplyParameters: &models.ReplyParameters{
			MessageID: replyToMsg.ID,
		},
		ChatID:    replyToMsg.Chat.ID,
		Text:      filterText(s),
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		msg, err = telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ReplyParameters: &models.ReplyParameters{
				MessageID: replyToMsg.ID,
			},
			ChatID: replyToMsg.Chat.ID,
			Text:   filterText(s),
		})
		if err != nil {
			fmt.Println("  reply send error:", err)
			msg = replyToMsg
		}
	}
	return
}

func editReplyToMessage(ctx context.Context, replyMsg *models.Message, s string) (msg *models.Message, err error) {
	msg, err = telegramBot.EditMessageText(ctx, &bot.EditMessageTextParams{
		MessageID: replyMsg.ID,
		ChatID:    replyMsg.Chat.ID,
		Text:      filterText(s),
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		msg, err = telegramBot.EditMessageText(ctx, &bot.EditMessageTextParams{
			MessageID: replyMsg.ID,
			ChatID:    replyMsg.Chat.ID,
			Text:      filterText(s),
		})
		if err != nil {
			fmt.Println("  reply edit error:", err)
			msg = replyMsg
		}
	}
	return
}

func deleteMessage(ctx context.Context, msg *models.Message) (success bool, err error) {
	if msg == nil {
		return
	}
	success, err = telegramBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		fmt.Println("  delete error:", err)
		success = false
	}
	return
}

func sendChatActionTyping(ctx context.Context, chatID int64) {
	action := bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	}

	_, err := telegramBot.SendChatAction(ctx, &action)
	if err != nil {
		fmt.Println("  send chat action error:", err)
	}
}

func sendTextToAdmins(ctx context.Context, s string) {
	for _, chatID := range params.AdminUserIDs {
		_, _ = sendMessage(ctx, chatID, s)
	}
}

func handleMessage(ctx context.Context, update *models.Update) {
	fmt.Print("msg from ", update.Message.From.Username, "#", update.Message.From.ID, ": ", update.Message.Text, "\n")

	if update.Message.Chat.ID >= 0 { // From user?
		if !slices.Contains(params.AllowedUserIDs, update.Message.From.ID) {
			fmt.Println("  user not allowed, ignoring")
			return
		}
	} else { // From group ?
		fmt.Print("  msg from group #", update.Message.Chat.ID)
		if !slices.Contains(params.AllowedGroupIDs, update.Message.Chat.ID) {
			fmt.Println(", group not allowed, ignoring")
			return
		}
		fmt.Println()
	}

	// Check if message is a command.
	if update.Message.Text[0] == '/' || update.Message.Text[0] == '!' {
		cmd := strings.Split(update.Message.Text, " ")[0]
		if strings.Contains(cmd, "@") {
			cmd = strings.Split(cmd, "@")[0]
		}
		update.Message.Text = strings.TrimPrefix(update.Message.Text, cmd+" ")
		update.Message.Text = strings.TrimPrefix(update.Message.Text, cmd)
		cmdChar := string(cmd[0])
		cmd = cmd[1:] // Cutting the command character.
		switch cmd {
		case params.ChatCmd:
			fmt.Println("  interpreting as chat cmd")
			cmdHandler.Chat(ctx, update.Message)
			return
		case "dsbalance":
			fmt.Println("  interpreting as cmd dsbalance")
			cmdHandler.Balance(ctx, update.Message, cmdChar)
			return
		case "dshelp":
			fmt.Println("  interpreting as cmd dshelp")
			cmdHandler.Help(ctx, update.Message, cmdChar)
			return
		case "start":
			fmt.Println("  interpreting as cmd start")
			if update.Message.Chat.ID >= 0 { // From user?
				_, _ = sendReplyToMessage(ctx, update.Message, "🤖 Welcome! This is the DeepSeek Telegram Bot\n\n"+
					"More info: https://github.com/nonoo/deepseek-telegram-bot")
			}
			return
		default:
			fmt.Println("  invalid cmd")
			if update.Message.Chat.ID >= 0 {
				_, _ = sendReplyToMessage(ctx, update.Message, errorStr+": invalid command")
			}
			return
		}
	}

	if update.Message.Chat.ID >= 0 || (update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == telegramBot.ID()) {
		cmdHandler.Chat(ctx, update.Message)
	}
}

func telegramBotUpdateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if update.Message.Text != "" {
		handleMessage(ctx, update)
	}
}

func main() {
	fmt.Println("deepseek-telegram-bot starting...")

	if err := params.Init(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	dsClient = deepseek.NewClient(params.DSAPIKey)

	var cancel context.CancelFunc
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(telegramBotUpdateHandler),
	}

	var err error
	telegramBot, err = bot.New(params.BotToken, opts...)
	if nil != err {
		panic(fmt.Sprint("can't init telegram bot: ", err))
	}

	sendTextToAdmins(ctx, "🤖 Bot started")

	telegramBot.Start(ctx)
}
