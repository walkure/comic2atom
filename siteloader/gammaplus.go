package siteloader

import (
	"fmt"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func gammaPlusFeed(target *url.URL) (string, *feeds.Feed, error) {
	doc, err := fetchDocument(target)
	if err != nil {
		return "", nil, fmt.Errorf("gammaplus:FetchErr:%w", err)
	}

	title := doc.Find("#top > div > article > section:nth-child(1) > div > ul > li:nth-child(1)").Text()
	if title == "" {
		return "", nil, fmt.Errorf("gammaplus:title not found")
	}

	author := doc.Find("#top > div > article > section:nth-child(1) > div > ul > li:nth-child(2)").Text()
	if author == "" {
		return "", nil, fmt.Errorf("gammaplus:author not found")
	}
	desc := trimDescription(doc.Find("#top > div > article > section:nth-child(3) > div > div.detail__area > div:nth-child(1) > p:nth-child(3)").Text())

	episodes := doc.Find("div.read__outer")

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}

	walkEpisode := func(i int, s *goquery.Selection) {
		title := trimDescription(s.Find("li.episode").Text())
		href, _ := s.Find("a").Attr("href")

		uri, _ := resolveRelativeURI(target, href)

		feed.Items = append(feed.Items, &feeds.Item{
			Title: title,
			Link:  &feeds.Link{Href: uri},
			Id:    generateHashedHex(uri),
		})
	}

	episodes.Each(walkEpisode)

	if len(feed.Items) == 0 {
		return "", nil, fmt.Errorf("gammaplus:no episode entry")
	}

	return "gammaplus_" + escapePath(target.Path), feed, nil
}
