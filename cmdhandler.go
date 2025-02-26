package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/go-telegram/bot/models"
)

const minReplyIntervalPrivateChat = time.Second
const minReplyIntervalGroupChat = 3 * time.Second

type cmdHandlerType struct {
	dsMsgHistory map[int64][]deepseek.ChatCompletionMessage
}

func (c *cmdHandlerType) reply(ctx context.Context, msg *models.Message, text string) (replyMsg *models.Message, err error) {
	if msg == nil || msg.Chat.ID >= 0 {
		return sendMessage(ctx, msg.Chat.ID, text)
	}
	return sendReplyToMessage(ctx, msg, text)
}

func (c *cmdHandlerType) editReply(ctx context.Context, msg *models.Message, replyMsg *models.Message, text string) (replyMessage *models.Message, err error) {
	if replyMsg == nil || msg == nil {
		return c.reply(ctx, msg, text)
	}

	return editReplyToMessage(ctx, replyMsg, text)
}

func (c *cmdHandlerType) Chat(ctx context.Context, msg *models.Message) {
	if c.dsMsgHistory == nil {
		c.dsMsgHistory = make(map[int64][]deepseek.ChatCompletionMessage)
	}

	request := &deepseek.StreamChatCompletionRequest{
		Model:       deepseek.DeepSeekChat,
		Temperature: float32(params.DSTemperature), // https://api-docs.deepseek.com/quick_start/parameter_settings
		MaxTokens:   params.DSMaxReplyTokens,
		Messages: []deepseek.ChatCompletionMessage{{
			Role:    constants.ChatMessageRoleSystem,
			Content: params.DSInitialPrompt,
		}},
		Stream: true,
	}

	request.Messages = append(request.Messages, c.dsMsgHistory[msg.Chat.ID]...)

	if msg.ReplyToMessage != nil {
		request.Messages = append(request.Messages, deepseek.ChatCompletionMessage{
			Role:    constants.ChatMessageRoleAssistant,
			Content: msg.ReplyToMessage.Text,
		})
	}

	request.Messages = append(request.Messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: msg.Text,
	})

	stream, err := dsClient.CreateChatCompletionStream(ctx, request)
	if err != nil {
		fmt.Println("    DS CreateChatCompletionStream error:", err)
		_, _ = c.reply(ctx, msg, errorStr+": "+err.Error())
		return
	}

	sendChatActionTyping(ctx, msg.Chat.ID)

	lastReplyEditAt := time.Now()
	minReplyInterval := minReplyIntervalPrivateChat
	if msg.Chat.ID < 0 {
		minReplyInterval = minReplyIntervalGroupChat
	}
	var replyMsg *models.Message
	var text string
	var lastsenttext string
	defer stream.Close()
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("    DS Recv error:", err)
			break
		}
		for _, choice := range response.Choices {
			text += choice.Delta.Content

			if time.Since(lastReplyEditAt) > minReplyInterval && text != "" && text != lastsenttext {
				replyMsg, _ = c.editReply(ctx, msg, replyMsg, text)
				lastReplyEditAt = time.Now()
				lastsenttext = text
			}
		}
	}

	if time.Since(lastReplyEditAt) < minReplyInterval && text != lastsenttext {
		time.Sleep(minReplyInterval - time.Since(lastReplyEditAt))
	}

	fmt.Println("    DS reply:", text)
	msg, err = c.editReply(ctx, msg, replyMsg, text)
	if err != nil {
		_, _ = deleteMessage(ctx, msg)
		_, _ = sendMessage(ctx, msg.Chat.ID, text)
	}

	if msg.ReplyToMessage != nil {
		c.dsMsgHistory[msg.Chat.ID] = append(c.dsMsgHistory[msg.Chat.ID], deepseek.ChatCompletionMessage{
			Role:    constants.ChatMessageRoleAssistant,
			Content: msg.ReplyToMessage.Text,
		})
	}
	c.dsMsgHistory[msg.Chat.ID] = append(c.dsMsgHistory[msg.Chat.ID], deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: msg.Text,
	})
	c.dsMsgHistory[msg.Chat.ID] = append(c.dsMsgHistory[msg.Chat.ID], deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleAssistant,
		Content: text,
	})
	for len(c.dsMsgHistory[msg.Chat.ID]) > params.DSHistorySize {
		c.dsMsgHistory[msg.Chat.ID] = c.dsMsgHistory[msg.Chat.ID][1:]
	}

	fmt.Println("    DS message history:")
	for i, msg := range c.dsMsgHistory[msg.Chat.ID] {
		fmt.Printf("    %d: %+v\n", i, msg)
	}
}

func (c *cmdHandlerType) Balance(ctx context.Context, msg *models.Message, cmdChar string) {
	balance, err := deepseek.GetBalance(dsClient, ctx)
	if err != nil {
		fmt.Println("    DS GetBalance error:", err)
		_, _ = c.reply(ctx, msg, errorStr+": "+err.Error())
		return
	}

	if balance == nil || len(balance.BalanceInfos) == 0 {
		_, _ = c.reply(ctx, msg, errorStr+": balance not available")
		return
	}

	replyText := fmt.Sprint("💰 ", balance.BalanceInfos[0].TotalBalance, " ", balance.BalanceInfos[0].Currency)
	fmt.Println("    DS reply:", replyText)
	_, _ = c.reply(ctx, msg, replyText)
}

func (c *cmdHandlerType) Help(ctx context.Context, msg *models.Message, cmdChar string) {
	_, _ = sendReplyToMessage(ctx, msg, "🤖 DeepSeek Telegram Bot\n\n"+
		"Available commands:\n\n"+
		cmdChar+params.ChatCmd+" - send chat message\n"+
		cmdChar+"dsbalance - show balance\n"+
		cmdChar+"dshelp - show this help\n\n"+
		"For more information see https://github.com/nonoo/deepseek-telegram-bot")
}
