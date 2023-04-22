package satoshiturk

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

type TelegramBot struct {
	Token   string
	ChatID  string
	BaseURL string
}

func NewTelegramBot(token, chatID string) *TelegramBot {
	return &TelegramBot{
		Token:   token,
		ChatID:  chatID,
		BaseURL: fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
	}
}

func (bot *TelegramBot) SendMessage(text string, options map[string]string) (blockResponse, error) {
	client := resty.New()

	data := map[string]string{
		"chat_id": bot.ChatID,
		"text":    text,
	}

	for k, v := range options {
		data[k] = v
	}

	response, err := client.R().SetFormData(data).Post(bot.BaseURL)
	if err != nil {
		log.Fatal(response)
	}

	return blockResponse{}, nil
}
