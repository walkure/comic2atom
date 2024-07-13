package siteloader

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func storiaFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {
	doc, metadata, err := fetchDocument(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("storia:FetchErr:%w", err)
	}

	title := doc.Find("#top > div > article > section:nth-child(1) > div > ul > li:nth-child(1)").Text()
	if title == "" {
		return "", nil, metadata, fmt.Errorf("storia:title not found")
	}

	author := doc.Find("#top > div > article > section:nth-child(1) > div > ul > li:nth-child(2)").Text()
	if author == "" {
		return "", nil, metadata, fmt.Errorf("storia:author not found")
	}
	desc := trimDescription(doc.Find("#top > div > article > section:nth-child(2) > div > div.detail__area > div:nth-child(1) > p").Text())

	episodes := doc.Find("div.read__area")

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}
	walkEpisode := func(i int, s *goquery.Selection) {
		title := s.Find("a > ul > li.read__detail > ul > li.episode").Text()
		href, _ := s.Find("a").Attr("href")
		img, _ := s.Find("a > ul > li.thumb > img").Attr("src")
		uri, _ := resolveRelativeURI(target, href)
		thumb, _ := resolveRelativeURI(target, img)

		feed.Items = append(feed.Items, &feeds.Item{
			Title:     title,
			Link:      &feeds.Link{Href: uri},
			Id:        generateHashedHex(uri),
			Enclosure: &feeds.Enclosure{Url: thumb},
		})
	}

	episodes.Find("div.read__outer").Each(walkEpisode)

	if len(feed.Items) == 0 {
		return "", nil, metadata, fmt.Errorf("storia:no episode entry")
	}

	return "storia_" + escapePath(target.Path), feed, metadata, nil
}
