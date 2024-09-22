package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"telegrambot/services/greeting"
	"telegrambot/services/rss"
	"telegrambot/services/treadsrepository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/mwazovzky/assistant"
	openaiclient "github.com/mwazovzky/assistant/http/client"
)

const botName = "Mike"

var allowedBots []string

func main() {
	bot := initBot()
	ai := initAssistant()
	allowedBots = []string{
		os.Getenv("BOT_CHAT_ID"),
		os.Getenv("TEST_CHAT_ID"),
		os.Getenv("FAMILY_CHAT_ID"),
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handle(ai, bot, update.Message)
		}
	}
}

func initAssistant() *openai.Assistant {
	url := "https://api.openai.com/v1/chat/completions"
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openaiclient.NewOpenAiClient(url, apiKey)
	tr := treadsrepository.NewThreadRepository()
	role := "You are assistant."

	return openai.NewAssistant(role, client, tr)
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

func handle(ai *openai.Assistant, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	log.Printf("Incoming message: chat_id: %d, from: %s, text: %s\n", msg.Chat.ID, msg.From.UserName, msg.Text)

	switch msg.Text {
	case "/start":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "Hello, human. AI welcomes you. What can I do for you?")
		return
	case "/help":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "Hello, human. AI welcomes you. Having a bad day? How can I help you?")
		return
	case "/habr":
		handleRSS(bot, msg)
		return
	case "heart":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "❤️")
		return
	case "like":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "👍")
		return
	case "ghost":
		sendReply(bot, msg.Chat.ID, msg.MessageID, "👻")
		return
	}

	// greetings response
	if greeting.ContainsGreeting(strings.ToLower(msg.Text)) {
		text := fmt.Sprintf("Привет, %s!", msg.From.FirstName)
		sendReply(bot, msg.Chat.ID, msg.MessageID, text)
		return
	}

	// send reactions
	emoji, ok := getReaction(msg.From.UserName)
	if ok {
		sendReaction(bot, msg.Chat.ID, msg.MessageID, emoji)
	}

	if slices.Contains(allowedBots, strconv.FormatInt(msg.Chat.ID, 10)) && strings.HasPrefix(msg.Text, botName) {
		handleQuestion(ai, bot, msg)
	}
}

func handleQuestion(ai *openai.Assistant, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := strings.TrimLeft(strings.TrimPrefix(msg.Text, "Mike"), "!, ")

	user := msg.From.UserName
	_, err := ai.GetThread(user)
	if err != nil {
		err = ai.CreateThread(user)
		if err != nil {
			log.Println("can not get or create thread, error:", err)
			return
		}
	}

	res, err := ai.Post(user, text)
	if err != nil {
		log.Println("can not post a question, error:", err)
		return
	}

	sendReply(bot, msg.Chat.ID, msg.MessageID, res)
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
