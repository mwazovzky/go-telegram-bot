package rss

import (
	"encoding/xml"
	"io"
	"net/http"
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

func GetNews(source string) (*RSS, error) {
	url := rss[source]
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
