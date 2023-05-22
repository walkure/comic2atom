package siteloader

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func kakuyomuFeed(target *url.URL) (string, *feeds.Feed, error) {
	doc, err := fetchDocument(target)
	if err != nil {
		return "", nil, fmt.Errorf("kakuyomu:FetchErr:%w", err)
	}

	title := doc.Find("#workTitle > a").Text()
	if title == "" {
		return "", nil, errors.New("kakuyomu:title not found")
	}

	author := doc.Find("#workAuthor-activityName > a").Text()
	if title == "" {
		return "", nil, errors.New("kakuyomu:author not found")
	}

	desc := trimDescription(doc.Find("#introduction").Text())

	updated, err := parseDatetimeEntity(doc.Find("p.widget-toc-date > time"))
	if err != nil {
		return "", nil, fmt.Errorf("kakuyomu:toc-date parse error %w", err)
	}

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Updated:     updated,
	}

	doc.Find("#table-of-contents > section > div > ol > li").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("span").Text())
		if title == "" {
			return
		}

		published, err := parseDatetimeEntity(s.Find("time"))
		if err != nil {
			return
		}

		href, ok := s.Find("a").Attr("href")
		if !ok {
			return
		}

		uri, _ := resolveRelativeURI(target, href)

		feed.Items = append(feed.Items, &feeds.Item{
			Title:   title,
			Link:    &feeds.Link{Href: uri},
			Id:      generateHashedHex(uri),
			Created: published,
		})
	})

	if len(feed.Items) == 0 {
		return "", nil, fmt.Errorf("kakuyomu:no episode entry")
	}

	return "kakuyomu_" + escapePath(target.Path), feed, nil
}

func parseDatetimeEntity(datetime *goquery.Selection) (time.Time, error) {

	dtText, ok := datetime.Attr("datetime")
	if !ok {
		return time.Time{}, errors.New("datetime not found")
	}

	tm, err := time.Parse(time.RFC3339, dtText)
	if err != nil {
		return time.Time{}, fmt.Errorf("datetime cannot parsed: %w", err)
	}

	return tm, nil
}
