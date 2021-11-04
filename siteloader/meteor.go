package siteloader

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func meteorFeed(target *url.URL) (string, *feeds.Feed, error) {
	doc, err := fetchDocument(target)
	if err != nil {
		return "", nil, fmt.Errorf("meteor:FetchErr:%w", err)
	}

	title := strings.TrimSpace(doc.Find("#contents > div.h2_area > h2 > div").Text())
	if title == "" {
		return "", nil, fmt.Errorf("meteor:title not found")
	}
	author := getTrimmedAuthor(doc.Find("#contents > div.work_author_intro.container-fluid > div > div.work_author_intro_txt_box > div.work_author_intro_name").Text())
	if author == "" {
		return "", nil, fmt.Errorf("meteor:author not found")
	}

	desc := trimDescription(doc.Find("#contents > div.work_story.container-fluid > div").Text())
	if desc == "" {
		return "", nil, fmt.Errorf("meteor:desc not found")
	}

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}

	episodes := doc.Find("#contents > div.work_episode > div.work_episode_box > div")

	walkEpisode := func(s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("div.work_episode_txt.d-table-cell").Text())
		uri, exist := s.Find("a").Attr("href")
		if exist {
			feed.Items = append(feed.Items, &feeds.Item{
				Title: title,
				Link:  &feeds.Link{Href: uri},
				Id:    generateHashedHex(uri),
			})
		}
	}

	episodes.Each(func(i int, s *goquery.Selection) {
		if s.HasClass("work_episode_table") {
			walkEpisode(s)
			return
		}

		if s.HasClass("moreEpi") {
			s.Find("div").Each(func(i int, s *goquery.Selection) {
				if s.HasClass("work_episode_table") {
					walkEpisode(s)
				}
			})
			return
		}

		if s.HasClass("episode_more_first") {
			walkEpisode(s)
			return
		}
	})

	if len(feed.Items) == 0 {
		return "", nil, fmt.Errorf("meteor:no episode entry")
	}

	return "meteor_" + escapePath(target.Path), feed, nil
}

func getTrimmedAuthor(author string) string {
	authorRune := []rune(strings.TrimSpace(author))
	trimmedAuthor := strings.TrimSpace(string(authorRune[2:]))
	return string([]rune(trimmedAuthor)[1:])
}
