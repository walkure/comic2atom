package siteloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func newtypeFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("newtype:failure to load Asia/Tokyo timezone: %w", err)
	}

	target, err = sanitizeComicWalkerURL(target)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("newtype:URLSanitizeErr:%w", err)

	}

	doc, metadata, err := fetchDocument(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("newtype:FetchErr:%w", err)
	}

	title := strings.TrimSpace(doc.Find("h1.contents__ttl").Text())
	if title == "" {
		return "", nil, metadata, fmt.Errorf("newtype:title not found")
	}

	author := strings.TrimSpace(doc.Find("div.contents__info").Text())
	if author == "" {
		return "", nil, metadata, fmt.Errorf("newtype:author not found")
	}

	desc := trimDescription(doc.Find("div.contents__txt--desc").Text())
	if desc == "" {
		return "", nil, metadata, fmt.Errorf("newtype:desc not found")
	}

	//fmt.Printf("newtype: title=[%s] author=[%s] desc=[%s]\n", title, author, desc)

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}

	feedName := "newtype_" + escapePath(target.Path)

	target.Path = path.Join(target.Path, "/more/1/Dsc")

	body, metadata, err := getHttpBody(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("newtype:FetchErr:%w", err)
	}

	var pagedata struct {
		HTML string `json:"html"`
		Next int    `json:"next"`
	}

	if err := json.Unmarshal(body, &pagedata); err != nil {
		return "", nil, metadata, fmt.Errorf("newtype:JSONParseErr:%w", err)
	}

	htmlReader := strings.NewReader(pagedata.HTML)
	doc, err = goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		return "", feed, metadata, fmt.Errorf("cannot create goquery document:%w", err)
	}

	doc.Find("li > a").Each(func(i int, s *goquery.Selection) {

		title := s.Find("h2.detail__txt--ttl-sub").Text()
		href, _ := s.Attr("href")
		img, _ := s.Find("img").Attr("src")

		caution := s.Find("div.detail__txt--caution").Text()
		if caution != "" {
			title = title + " " + caution
		}

		uri, _ := resolveRelativeURI(target, href)
		thumb, _ := resolveRelativeURI(target, img)

		dateStr := strings.Split(s.Find("div.detail__txt--date").Text(), " ")[0]
		date, err := time.ParseInLocation("2006/1/2", dateStr, loc)
		if err != nil {
			fmt.Printf("newtype: date parse error:%v\n", err)
			date = time.Now()
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Title:     title,
			Link:      &feeds.Link{Href: uri},
			Id:        generateHashedHex(uri + title),
			Created:   date,
			Enclosure: &feeds.Enclosure{Url: thumb},
		})
	})

	return feedName, feed, metadata, nil
}
