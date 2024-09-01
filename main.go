package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"telegrambot/services/greeting"
	"telegrambot/services/rss"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot := initBot()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handle(bot, update.Message)
		}
	}

}

func initBot() *tgbotapi.BotAPI {
	botToken := os.Getenv("TELEGRAM_HTTP_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return bot
}

func handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	log.Printf("Incoming message: chat_id: %d, from: %s, text: %s\n", msg.Chat.ID, msg.From.UserName, msg.Text)

	switch msg.Text {
	case "/start":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "Hello, human. AI welcomes you. What can I do for you?")
	case "/help":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "Hello, human. AI welcomes you. Having a bad day? How can I help you?")
	case "/habr":
		handleRSS(bot, msg)
	case "test":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "Testing")
	case "bye":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "Goodbye. Have a nice day!")
	case "heart":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "❤️")
	case "like":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "👍")
	case "ghost":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "👻")
	}

	if greeting.ContainsGreeting(strings.ToLower(msg.Text)) {
		text := fmt.Sprintf("Привет, %s!", msg.From.FirstName)
		sendReply(bot, msg.Chat.ID, msg.MessageID, text)
	}

	emoji, ok := getReaction(msg.From.UserName)
	if ok {
		sendReaction(bot, msg.Chat.ID, msg.MessageID, emoji)
	}
}

func handleRSS(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	feed, err := rss.GetNews("habr")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Sorry, Can not load new at the moment"))
	}
	for _, item := range feed.Items {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, item.URL+"\n"+item.Title))
	}
}

func sendReply(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = messageID
	bot.Send(msg)
	log.Printf("Outgoing message [reply]: ChatID: %d, Text: %s, ReplyToMessageID: %d", msg.ChatID, msg.Text, messageID)
}

func getReaction(username string) (string, bool) {
	symbols := []string{"💋", "❤️", "👀", "👀", "👀"}
	names := os.Getenv("TEAM")
	m := map[string]string{}
	for i, name := range strings.Split(names, ",") {
		m[name] = symbols[i]
	}
	value, ok := m[username]

	return value, ok
}

func sendReaction(bot *tgbotapi.BotAPI, chatID int64, messageID int, emoji string) {
	params := tgbotapi.Params{}
	params.AddNonZero64("chat_id", chatID)
	params.AddNonZero("message_id", messageID)
	reaction := fmt.Sprintf("[{\"type\":\"emoji\",\"emoji\":\"%s\"}]", emoji)
	params.AddNonEmpty("reaction", reaction)
	_, err := bot.MakeRequest("setMessageReaction", params)
	if err != nil {
		log.Println("ERROR", err)
	}
	log.Printf("Outgoing message [reaction]: ChatID: %d, Text: %s, ReplyToMessageID: %d", chatID, emoji, messageID)
}
