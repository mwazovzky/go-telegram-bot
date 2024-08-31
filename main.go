package main

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var rss = map[string]string{
	"habr": "https://habrahabr.ru/rss/best/",
}

type RSS struct {
	Items []Item `xml:"channel>item"`
}

type Item struct {
	URL   string `xml:"guid"`
	Title string `xml:"title"`
}

func getNews(url string) (*RSS, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	rss := new(RSS)
	err = xml.Unmarshal(body, rss)
	if err != nil {
		return nil, err
	}

	return rss, nil
}

func main() {
	botToken := os.Getenv("TELEGRAM_HTTP_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			im := update.Message
			log.Printf("Incoming message: chat_id: %d, from: %s, text: %s", im.Chat.ID, im.From.UserName, im.Text)
			handle(bot, update.Message)
		}
	}
}

var greetings = []string{
	"доброе утро",
	"доброе день",
	"доброе вечер",
	"утро доброе",
	"привет",
	"good morning",
	"hello",
	"hi",
}

func handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Text {
	case "/start":
		send(bot, msg.Chat.ID, msg.MessageID, "Hello, human. AI welcomes you. What can I do for you?")
	case "/help":
		send(bot, msg.Chat.ID, msg.MessageID, "Hello, human. AI welcomes you. Having a bad day? How can I help you?")
	case "/habr":
		handleRSS(bot, msg)
	case "test":
		send(bot, msg.Chat.ID, msg.MessageID, "Testing")
	case "bye":
		send(bot, msg.Chat.ID, msg.MessageID, "Goodbye. Have a nice day!")
	default:
		handleGeeting(bot, msg)
	}
}

func handleGeeting(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if containsGreeting(strings.ToLower(msg.Text)) {
		send(bot, msg.Chat.ID, msg.MessageID, "Привет, человеки!")
	}
}

func containsGreeting(text string) bool {
	str := strings.ToLower(text)

	for _, greeting := range greetings {
		if strings.Contains(str, greeting) {
			return true
		}
	}

	return false
}

func handleRSS(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	url := rss["habr"]
	feed, err := getNews(url)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Sorry, Can not load new at the moment"))
	}
	for _, item := range feed.Items {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, item.URL+"\n"+item.Title))
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = messageID
	bot.Send(msg)
	log.Printf("Ougoing message: chat_id: %d, text: %s", msg.ChatID, msg.Text)
}
