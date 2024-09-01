package greeting

import "strings"

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

func ContainsGreeting(text string) bool {
	str := strings.ToLower(text)

	for _, greeting := range greetings {
		if strings.Contains(str, greeting) {
			return true
		}
	}

	return false
}
