package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type paramsType struct {
	DSAPIKey string
	BotToken string

	DSInitialPrompt  string
	DSTemperature    float64
	DSMaxReplyTokens int
	DSHistorySize    int

	AllowedUserIDs  []int64
	AdminUserIDs    []int64
	AllowedGroupIDs []int64
}

var params paramsType

func (p *paramsType) Init() error {
	flag.StringVar(&p.DSAPIKey, "ds-api-key", "", "deepseek api key")
	flag.StringVar(&p.BotToken, "bot-token", "", "telegram bot token")
	flag.StringVar(&p.DSInitialPrompt, "ds-initial-prompt", "", "deepseek initial prompt")
	DSMaxReplyTokensFlag := flag.Int("ds-max-reply-tokens", math.MaxInt, "deepseek max reply tokens")
	DSHistorySizeFlag := flag.Int("ds-history-size", math.MaxInt, "deepseek message history size")
	flag.Float64Var(&p.DSTemperature, "ds-temperature", math.MaxFloat64, "deepseek temperature")
	var allowedUserIDs string
	flag.StringVar(&allowedUserIDs, "allowed-user-ids", "", "allowed telegram user ids")
	var adminUserIDs string
	flag.StringVar(&adminUserIDs, "admin-user-ids", "", "admin telegram user ids")
	var allowedGroupIDs string
	flag.StringVar(&allowedGroupIDs, "allowed-group-ids", "", "allowed telegram group ids")
	flag.Parse()

	if p.DSAPIKey == "" {
		p.DSAPIKey = os.Getenv("DS_API_KEY")
	}
	if p.DSAPIKey == "" {
		return fmt.Errorf("ds api key not set")
	}

	if p.DSInitialPrompt == "" {
		p.DSInitialPrompt = os.Getenv("DS_INITIAL_PROMPT")
	}

	if p.DSTemperature == math.MaxFloat64 {
		temperatureStr := os.Getenv("DS_TEMPERATURE")
		if temperatureStr != "" {
			temperature, err := strconv.ParseFloat(temperatureStr, 64)
			if err != nil {
				return fmt.Errorf("invalid deepseek temperature: %s", temperatureStr)
			}
			p.DSTemperature = temperature
		} else {
			p.DSTemperature = 1.3
		}
	}

	if DSMaxReplyTokensFlag != nil && *DSMaxReplyTokensFlag != math.MaxInt {
		p.DSMaxReplyTokens = *DSMaxReplyTokensFlag
	} else {
		maxReplyTokensStr := os.Getenv("DS_MAX_REPLY_TOKENS")
		if maxReplyTokensStr != "" {
			maxReplyTokens, err := strconv.Atoi(maxReplyTokensStr)
			if err != nil {
				return fmt.Errorf("invalid deepseek max reply tokens: %s", maxReplyTokensStr)
			}
			p.DSMaxReplyTokens = maxReplyTokens
		} else {
			p.DSMaxReplyTokens = 2048
		}
	}

	if DSHistorySizeFlag != nil && *DSHistorySizeFlag != math.MaxInt {
		p.DSHistorySize = *DSHistorySizeFlag
	} else {
		historySizeStr := os.Getenv("DS_HISTORY_SIZE")
		if historySizeStr != "" {
			historySize, err := strconv.Atoi(historySizeStr)
			if err != nil {
				return fmt.Errorf("invalid deepseek history size: %s", historySizeStr)
			}
			p.DSHistorySize = historySize
		} else {
			p.DSHistorySize = 4
		}
	}

	if p.BotToken == "" {
		p.BotToken = os.Getenv("BOT_TOKEN")
	}
	if p.BotToken == "" {
		return fmt.Errorf("bot token not set")
	}

	if allowedUserIDs == "" {
		allowedUserIDs = os.Getenv("ALLOWED_USERIDS")
	}
	sa := strings.Split(allowedUserIDs, ",")
	for _, idStr := range sa {
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("allowed user ids contains invalid user ID: %s", idStr)
		}
		p.AllowedUserIDs = append(p.AllowedUserIDs, id)
	}

	if adminUserIDs == "" {
		adminUserIDs = os.Getenv("ADMIN_USERIDS")
	}
	sa = strings.Split(adminUserIDs, ",")
	for _, idStr := range sa {
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("admin ids contains invalid user ID: %s", idStr)
		}
		p.AdminUserIDs = append(p.AdminUserIDs, id)
		if !slices.Contains(p.AllowedUserIDs, id) {
			p.AllowedUserIDs = append(p.AllowedUserIDs, id)
		}
	}

	if allowedGroupIDs == "" {
		allowedGroupIDs = os.Getenv("ALLOWED_GROUPIDS")
	}
	sa = strings.Split(allowedGroupIDs, ",")
	for _, idStr := range sa {
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("allowed group ids contains invalid group ID: %s", idStr)
		}
		p.AllowedGroupIDs = append(p.AllowedGroupIDs, id)
	}

	return nil
}
