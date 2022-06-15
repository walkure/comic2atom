package siteloader

import (
	"fmt"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func valkyrieFeed(target *url.URL) (string, *feeds.Feed, error) {
	doc, err := fetchDocument(target)
	if err != nil {
		return "", nil, fmt.Errorf("valkyrie:FetchErr:%w", err)
	}

	title := doc.Find("title").Text()
	if title == "" {
		return "", nil, fmt.Errorf("valkyrie:title not found")
	}

	author := doc.Find("#writer > p").Text()
	if author == "" {
		return "", nil, fmt.Errorf("valkyrie:author not found")
	}
	author = trimDescription(author)

	desc := trimDescription(doc.Find("#bg > main > div > div.t_box > p").Text())
	desc = trimDescription(desc)

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: target.String()},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Created:     time.Now(),
	}

	title = doc.Find("#new_story > div > p.title").Text()
	href, _ := doc.Find("#new_story > div > a").Attr("href")
	img, _ := doc.Find("#new_story > figure > img").Attr("src")
	img, _ = resolveRelativeURI(target, img)

	feed.Items = append(feed.Items, &feeds.Item{
		Title:     title,
		Link:      &feeds.Link{Href: href},
		Id:        generateHashedHex(href),
		Enclosure: &feeds.Enclosure{Url: img},
	})

	doc.Find("#back_number > div > div").Each(func(i int, s *goquery.Selection) {
		title := s.Find("p.title").Text()
		href, _ := s.Find("div > a").Attr("href")
		img, _ := s.Find("figure > img").Attr("src")
		img, _ = resolveRelativeURI(target, img)

		feed.Items = append(feed.Items, &feeds.Item{
			Title:     title,
			Link:      &feeds.Link{Href: href},
			Enclosure: &feeds.Enclosure{Url: img},
			Id:        generateHashedHex(href),
		})
	})

	if len(feed.Items) == 0 {
		return "", nil, fmt.Errorf("valkyrie:no episode entry")
	}

	return "valkyrie_" + escapePath(target.Path), feed, nil
}
