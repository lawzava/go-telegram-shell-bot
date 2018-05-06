package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// TelegramAPIKey defines the API key of bot, received from @BotFather in Telegram
var TelegramAPIKey = os.Getenv("TELEGRAM_API_KEY")

// TelegramDebugMode enables debug options
var TelegramDebugMode = os.Getenv("TELEGRAM_DEBUG")

func main() {
	bot, err := tgbotapi.NewBotAPI(TelegramAPIKey)
	if err != nil {
		log.Panic(err)
	}

	if TelegramDebugMode == "true" {
		bot.Debug = true
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if TelegramDebugMode == "true" {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		}

		go sendData(update.Message.Chat.ID, update.Message.Text, bot)
	}
}

func sendData(chatID int64, command string, bot *tgbotapi.BotAPI) {
	stdout, err := executeCommand(command)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprint(err))
		bot.Send(msg)
		return
	}

	var sentMessageID int
	var endString []byte
	p := make([]byte, 156)
	for {
		n, err := stdout.Read(p)
		if err == io.EOF {
			if sentMessageID != 0 {
				msgEdit := tgbotapi.EditMessageTextConfig{
					BaseEdit: tgbotapi.BaseEdit{
						ChatID:    chatID,
						MessageID: sentMessageID,
					},
					Text: string(endString),
				}
				if msgEdit.Text != "" {
					sentMessage, _ := bot.Send(msgEdit)
					if sentMessageID == 0 {
						sentMessageID = sentMessage.MessageID
					}
				}
				break
			}

			msg := tgbotapi.NewMessage(chatID, string(endString))
			if msg.Text != "" {
				sentMessage, _ := bot.Send(msg)
				if sentMessageID == 0 {
					sentMessageID = sentMessage.MessageID
				}
			}
			break
		}
		endString = append(endString, p[:n]...)
		if sentMessageID != 0 {
			msgEdit := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:    chatID,
					MessageID: sentMessageID,
				},
				Text: string(endString),
			}
			if msgEdit.Text != "" {
				sentMessage, _ := bot.Send(msgEdit)
				if sentMessageID == 0 {
					sentMessageID = sentMessage.MessageID
				}
			}
			continue
		}

		msg := tgbotapi.NewMessage(chatID, string(endString))
		if msg.Text != "" {
			sentMessage, _ := bot.Send(msg)
			if sentMessageID == 0 {
				sentMessageID = sentMessage.MessageID
			}
		}
	}
}

func executeCommand(command string) (io.ReadCloser, error) {
	parts := strings.Fields(command)
	cmd := exec.Command(parts[0])
	if len(parts) > 1 {
		cmd = exec.Command(parts[0], parts[1:]...)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error while executing command: %v", err)
	}
	cmd.Start()
	return stdout, nil
}
