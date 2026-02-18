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

func alphapolisMOFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {

	doc, metadata, err := fetchDocument(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("alphapolisMO:FetchErr:%w", err)
	}

	title := doc.Find("meta[property='og:title']").AttrOr("content", "タイトル不明")
	link, _ := doc.Find("link[rel='canonical']").Attr("href")

	description := doc.Find("meta[name='description']").AttrOr("content", "")

	var authors []string
	doc.Find(".author-label .mangaka").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		if name != "" {
			authors = append(authors, name)
		}
	})
	authorString := strings.Join(authors, " | ")

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: link},
		Description: description,
		Author:      &feeds.Author{Name: authorString},
		Created:     time.Now(),
		Id:          generateHashedHex(link),
	}

	doc.Find(".episode-unit").Each(func(i int, s *goquery.Selection) {

		freeNode := s.Find(".free")
		if freeNode.Length() == 0 {
			return // 無料でない場合はスキップ
		}

		eTitle := strings.TrimSpace(s.Find(".title").First().Text())
		eHref := ""
		if val, ok := s.Find("a.read-episode").Attr("href"); ok && val != "#" {
			eHref = val
		} else if val, ok := s.Find("a.read-comments").Attr("href"); ok {
			eHref = strings.Split(val, "?")[0]
		}

		eHref, err = resolveRelativeURI(target, eHref)
		if err != nil {
			return
		}

		updateTime := s.Find(".up-time").Text()
		limitText := strings.TrimSpace(freeNode.Text()) // 「2026.02.24まで無料」など
		thumbURL, _ := s.Find("img").Attr("data-src")
		if thumbURL == "" {
			thumbURL, _ = s.Find("img").Attr("src")
		}
		item := &feeds.Item{
			Title:       eTitle,
			Link:        &feeds.Link{Href: eHref},
			Description: fmt.Sprintf("%s（更新日: %s）", limitText, updateTime),
			Created:     parseAPMCDate(updateTime),
			Id:          generateHashedHex(eHref),
			Enclosure:   &feeds.Enclosure{Url: thumbURL},
		}
		feed.Items = append(feed.Items, item)
	})

	if len(feed.Items) == 0 {
		return "", nil, metadata, fmt.Errorf("alphapolisMO:no episode entry")
	}

	return "alphapolis_" + escapePath(target.Path), feed, metadata, nil

}

func parseAPMCDate(raw string) time.Time {
	clean := strings.ReplaceAll(raw, "更新", "")
	clean = strings.TrimSpace(clean)

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}
	}

	t, err := time.ParseInLocation("2006.01.02", clean, loc)
	if err != nil {
		return time.Now()
	}
	return t
}
