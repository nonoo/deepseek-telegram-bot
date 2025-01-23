package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/go-telegram/bot/models"
)

const minReplyInterval = time.Second

type cmdHandlerType struct {
	dsMsgHistory map[int64][]deepseek.ChatCompletionMessage
}

func (c *cmdHandlerType) reply(ctx context.Context, msg *models.Message, text string) *models.Message {
	if msg.Chat.ID >= 0 || msg == nil {
		return sendMessage(ctx, msg.Chat.ID, text)
	}
	return sendReplyToMessage(ctx, msg, text)
}

func (c *cmdHandlerType) editReply(ctx context.Context, msg *models.Message, replyMsg *models.Message, text string) *models.Message {
	if replyMsg == nil || msg == nil {
		return c.reply(ctx, msg, text)
	}

	return editReplyToMessage(ctx, replyMsg, text)
}

func (c *cmdHandlerType) DS(ctx context.Context, msg *models.Message) {
	if c.dsMsgHistory == nil {
		c.dsMsgHistory = make(map[int64][]deepseek.ChatCompletionMessage)
	}

	request := &deepseek.StreamChatCompletionRequest{
		Model:       deepseek.DeepSeekChat,
		Temperature: float32(params.DSTemperature), // https://api-docs.deepseek.com/quick_start/parameter_settings
		MaxTokens:   params.DSMaxReplyTokens,
		Messages: []deepseek.ChatCompletionMessage{{
			Role:    deepseek.ChatMessageRoleSystem,
			Content: params.DSInitialPrompt,
		}},
		Stream: true,
	}

	request.Messages = append(request.Messages, c.dsMsgHistory[msg.Chat.ID]...)

	if msg.ReplyToMessage != nil {
		request.Messages = append(request.Messages, deepseek.ChatCompletionMessage{
			Role:    deepseek.ChatMessageRoleAssistant,
			Content: msg.ReplyToMessage.Text,
		})
	}

	request.Messages = append(request.Messages, deepseek.ChatCompletionMessage{
		Role:    deepseek.ChatMessageRoleUser,
		Content: msg.Text,
	})

	stream, err := dsClient.CreateChatCompletionStream(ctx, request)
	if err != nil {
		fmt.Println("    DS CreateChatCompletionStream error:", err)
		c.reply(ctx, msg, errorStr+": "+err.Error())
		return
	}

	sendChatActionTyping(ctx, msg.Chat.ID)

	lastReplyEditAt := time.Now()
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

			if time.Since(lastReplyEditAt) > minReplyInterval && text != lastsenttext {
				replyMsg = c.editReply(ctx, msg, replyMsg, text)
				lastReplyEditAt = time.Now()
				lastsenttext = text
			}
		}
	}

	if time.Since(lastReplyEditAt) < minReplyInterval && text != lastsenttext {
		time.Sleep(minReplyInterval - time.Since(lastReplyEditAt))
	}

	fmt.Println("    DS reply:", text)
	c.editReply(ctx, msg, replyMsg, text)

	if msg.ReplyToMessage != nil {
		c.dsMsgHistory[msg.Chat.ID] = append(c.dsMsgHistory[msg.Chat.ID], deepseek.ChatCompletionMessage{
			Role:    deepseek.ChatMessageRoleAssistant,
			Content: msg.ReplyToMessage.Text,
		})
	}
	c.dsMsgHistory[msg.Chat.ID] = append(c.dsMsgHistory[msg.Chat.ID], deepseek.ChatCompletionMessage{
		Role:    deepseek.ChatMessageRoleUser,
		Content: msg.Text,
	})
	c.dsMsgHistory[msg.Chat.ID] = append(c.dsMsgHistory[msg.Chat.ID], deepseek.ChatCompletionMessage{
		Role:    deepseek.ChatMessageRoleAssistant,
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
		c.reply(ctx, msg, errorStr+": "+err.Error())
		return
	}

	if balance == nil || len(balance.BalanceInfos) == 0 {
		c.reply(ctx, msg, errorStr+": balance not available")
		return
	}

	replyText := fmt.Sprint("💰 ", balance.BalanceInfos[0].TotalBalance, " ", balance.BalanceInfos[0].Currency)
	fmt.Println("    DS reply:", replyText)
	c.reply(ctx, msg, replyText)
}

func (c *cmdHandlerType) Help(ctx context.Context, msg *models.Message, cmdChar string) {
	sendReplyToMessage(ctx, msg, "🤖 DeepSeek Telegram Bot\n\n"+
		"Available commands:\n\n"+
		cmdChar+"ds - send chat message\n"+
		cmdChar+"dsbalance - show balance\n"+
		cmdChar+"dshelp - show this help\n\n"+
		"For more information see https://github.com/nonoo/deepseek-telegram-bot")
}
