package siteloader

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func narouFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {
	doc, metadata, err := fetchDocument(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("storia:FetchErr:%w", err)
	}

	title := doc.Find("h1.p-novel__title").Text()
	if title == "" {
		return "", nil, metadata, fmt.Errorf("narou:title not found")
	}

	author := doc.Find("div.p-novel__author > a").Text()
	if author == "" {
		return "", nil, metadata, fmt.Errorf("narou:author not found")
	}

	desc := doc.Find("div.p-novel__summary").Text()
	if desc == "" {
		return "", nil, metadata, fmt.Errorf("narou:description not found")
	}

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: trimDescription(desc),
		Author:      &feeds.Author{Name: author},
	}

	chapter := ""
	eachError := error(nil)

	collectArticles := func(doc *goquery.Document) {
		doc.Find("div.p-eplist").Children().EachWithBreak(func(i int, s *goquery.Selection) bool {
			if s.Is("div.p-eplist__chapter-title") {
				chapter = s.Text()
				return true
			}

			if s.Is("div.p-eplist__sublist") {
				subject := s.Find("a.p-eplist__subtitle")
				subtitle := trimDescription(subject.Text())
				link, ok := subject.Attr("href")
				if !ok {
					eachError = errors.New("cannot find href")
					return false
				}
				href, err := resolveRelativeURI(target, link)
				if err != nil {
					eachError = fmt.Errorf("cannot parse URL:%w", err)
					return false
				}

				fulltitle := subtitle
				if chapter != "" {
					fulltitle = chapter + "/" + subtitle
				}

				it := &feeds.Item{
					Title: fulltitle,
					Link:  &feeds.Link{Href: href},
					Id:    generateHashedHex(href),
				}

				created := s.Find("div.p-eplist__update").Text()
				if created == "" {
					eachError = errors.New("cannot find created timestamp")
					return false
				}
				parsed, err := parseTimestamp(created)
				if err != nil {
					eachError = fmt.Errorf("cannot parse created[%s]:%w", created, err)
					return false
				}
				it.Created = parsed
				if parsed.After(feed.Updated) {
					feed.Updated = parsed
				}

				updated, ok := s.Find("div.p-eplist__update > span").Attr("title")
				if ok {
					parsed, err := parseTimestamp(updated)
					if err != nil {
						eachError = fmt.Errorf("cannot parse updated[%s]:%w", updated, err)
						return false
					}
					it.Updated = parsed
					if parsed.After(feed.Updated) {
						feed.Updated = parsed
					}
				}

				feed.Items = append(feed.Items, it)
			}

			return true
		})
	}

	for {
		collectArticles(doc)

		next, ok := doc.Find(`a[class="c-pager__item c-pager__item--next"]`).Attr("href")

		if !ok {
			break
		}

		nextURL, err := target.Parse(next)
		if err != nil {
			return "", nil, metadata, fmt.Errorf("narou:cannot parse next URL:%w", err)
		}

		// use latest metadata(etag and last-modified) for next request
		doc, metadata, err = fetchDocument(ctx, nextURL)
		if err != nil {
			return "", nil, metadata, fmt.Errorf("narou:Fetch(Next)Err:%w", err)
		}
	}

	if eachError != nil {
		return "", nil, metadata, fmt.Errorf("narou:%w", eachError)
	}

	if len(feed.Items) == 0 {
		return "", nil, metadata, fmt.Errorf("narou:no episode entry")
	}

	return "narou_" + escapePath(target.Path), feed, metadata, nil
}

func parseTimestamp(str string) (time.Time, error) {
	cleanup := trimDescription(str)
	// 1234567890123456
	// 2006/01/02 15:04
	if len(cleanup) < 16 {
		return time.Time{}, errors.New("time string is too short")
	}
	filtered := cleanup[:16]

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation("2006/01/02 15:04", filtered, loc)
}
