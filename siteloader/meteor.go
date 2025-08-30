package siteloader

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func meteorFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {
	doc, metadata, err := fetchDocument(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("meteor:FetchErr:%w", err)
	}

	title := strings.TrimSpace(doc.Find("body > main > h2").Text())
	if title == "" {
		return "", nil, metadata, fmt.Errorf("meteor:title not found")
	}
	author := getTrimmedAuthor(doc.Find("body > main > div.content-container > div.group-button-r2.mt-4.mb-5 > a").Text())
	if author == "" {
		return "", nil, metadata, fmt.Errorf("meteor:author not found")
	}

	desc := trimDescription(doc.Find("body > main > div.content-container > div.link-color.lh-lg.mx-2.mx-lg-0").Text())
	if desc == "" {
		return "", nil, metadata, fmt.Errorf("meteor:desc not found")
	}

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}

	episodes := doc.Find("div.episode-item")

	episodes.Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("div.fw-bold").Text())
		uri, exist := s.Find("a").Attr("href")
		if exist {
			feed.Items = append(feed.Items, &feeds.Item{
				Title: title,
				Link:  &feeds.Link{Href: uri},
				Id:    generateHashedHex(uri),
			})
		}
	})

	if len(feed.Items) == 0 {
		return "", nil, metadata, fmt.Errorf("meteor:no episode entry")
	}

	return "meteor_" + escapePath(target.Path), feed, metadata, nil
}

func getTrimmedAuthor(author string) string {
	authorRune := []rune(strings.TrimSpace(author))
	trimmedAuthor := strings.TrimSpace(string(authorRune[2:]))
	return string([]rune(trimmedAuthor)[1:])
}
