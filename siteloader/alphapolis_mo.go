package siteloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

// alphapolisMOData represents the JSON structure embedded in the page
type alphapolisMOData struct {
	Episodes []alphapolisMOEpisode `json:"episodes"`
}

type alphapolisMOEpisode struct {
	EpisodeNo    int                `json:"episodeNo"`
	URL          string             `json:"url"`
	ShortTitle   string             `json:"shortTitle"`
	MainTitle    string             `json:"mainTitle"`
	ThumbnailURL string             `json:"thumbnailUrl"`
	UpTime       string             `json:"upTime"`
	Rental       alphapolisMORental `json:"rental"`
}

type alphapolisMORental struct {
	IsFree     bool   `json:"isFree"`
	FreeExpire *int64 `json:"freeExpire"`
}

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

	// Find and parse JSON data from script tag
	var episodesData alphapolisMOData
	doc.Find("#app-official-manga-toc script[type='application/json']").Each(func(i int, s *goquery.Selection) {
		jsonText := s.Text()
		if err := json.Unmarshal([]byte(jsonText), &episodesData); err != nil {
			return
		}
	})

	if len(episodesData.Episodes) == 0 {
		return "", nil, metadata, fmt.Errorf("alphapolisMO:no episode data found")
	}

	// Process episodes from JSON
	for _, ep := range episodesData.Episodes {
		// Skip non-free episodes
		if !ep.Rental.IsFree {
			continue
		}

		eHref, err := resolveRelativeURI(target, ep.URL)
		if err != nil {
			continue
		}

		// Build description with free status
		description := ""
		if ep.UpTime != "" {
			description = fmt.Sprintf("更新日: %s", ep.UpTime)
		}
		if ep.Rental.FreeExpire != nil {
			expireTime := time.Unix(*ep.Rental.FreeExpire/1000, 0)
			description = fmt.Sprintf("%sまで無料 (%s)", expireTime.Format("2006.01.02"), description)
		} else {
			description = fmt.Sprintf("無料 (%s)", description)
		}

		item := &feeds.Item{
			Title:       ep.ShortTitle,
			Link:        &feeds.Link{Href: eHref},
			Description: description,
			Created:     parseAPMCDate(ep.UpTime),
			Id:          generateHashedHex(eHref),
			Enclosure:   &feeds.Enclosure{Url: ep.ThumbnailURL},
		}
		feed.Items = append(feed.Items, item)
	}

	if len(feed.Items) == 0 {
		return "", nil, metadata, fmt.Errorf("alphapolisMO:no free episode entry")
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
