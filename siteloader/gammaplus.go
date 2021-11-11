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

	title := doc.Find("#main_contents > article.main_area > section.work_main > div.col_work_name > h1").Text()
	if title == "" {
		return "", nil, fmt.Errorf("gammaplus:title not found")
	}

	author := doc.Find("#main_contents > article.main_area > section.work_main > div.col_work_name > div.author").Text()
	if author == "" {
		return "", nil, fmt.Errorf("gammaplus:author not found")
	}
	desc := trimDescription(doc.Find("#main_contents > article.main_area > section.work_main > p").Text())

	episodes := doc.Find("#main_contents > article.main_area > section.episode > div.box_episode")

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}

	walkEpisode := func(i int, s *goquery.Selection) {
		title := s.Find("div.episode_title").Text()
		caption := s.Find("div.episode_caption").Text()
		href, _ := s.Find("a").Attr("href")

		uri, _ := resolveRelativeURI(target, href)

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: uri},
			Description: caption,
			Id:          generateHashedHex(uri),
		})
	}

	episodes.Find("div.box_episode_L").Each(walkEpisode)
	episodes.Find("div.box_episode_M").Each(walkEpisode)

	if len(feed.Items) == 0 {
		return "", nil, fmt.Errorf("gammaplus:no episode entry")
	}

	return "gammaplus_" + escapePath(target.Path), feed, nil
}
