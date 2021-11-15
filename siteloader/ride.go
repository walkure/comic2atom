package siteloader

import (
	"fmt"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func rideFeed(target *url.URL) (string, *feeds.Feed, error) {
	doc, err := fetchDocument(target)
	if err != nil {
		return "", nil, fmt.Errorf("ride:FetchErr:%w", err)
	}

	title := doc.Find("body > div > main > div:nth-child(1) > div > div > div.p-detail-head__main > h1").Text()
	if title == "" {
		return "", nil, fmt.Errorf("ride:title not found")
	}

	author := doc.Find("body > div > main > div:nth-child(1) > div > div > div.p-detail-head__main > p").Text()
	if author == "" {
		return "", nil, fmt.Errorf("ride:author not found")
	}

	desc := trimDescription(doc.Find("body > div > main > div:nth-child(1) > div > p").Text())

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}

	doc.Find("body > div > main > div.c-section > div > div > div").Each(func(i int, s *goquery.Selection) {
		s.Find("ul > li.p-backnumber-d").Each(func(j int, s *goquery.Selection) {
			title := s.Find("strong > span").Text()
			href, _ := s.Find("span > a").Attr("href")
			feed.Items = append(feed.Items, &feeds.Item{
				Title: title,
				Link:  &feeds.Link{Href: href},
				Id:    generateHashedHex(href),
			})
		})
	})

	if len(feed.Items) == 0 {
		return "", nil, fmt.Errorf("ride:no episode entry")
	}

	return "ride_" + escapePath(target.Path), feed, nil
}
