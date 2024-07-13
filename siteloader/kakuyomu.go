package siteloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func getTimFromObj(i interface{}) (time.Time, error) {

	if i == nil {
		return time.Time{}, errors.New("date is nil")
	}

	t, ok := i.(string)
	if !ok {
		return time.Time{}, errors.New("invalid date type")
	}

	return time.Parse(time.RFC3339, t)
}

func kakuyomuFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {
	doc, metadata, err := fetchDocument(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("kakuyomu:FetchErr:%w", err)
	}

	script := doc.Find("script#__NEXT_DATA__").Text()
	if script == "" {
		return "", nil, metadata, errors.New("kakuyomu:__NEXT_DATA__ not found")
	}

	var kakuyomuNextData struct {
		Props struct {
			PageProps struct {
				ApolloState map[string]interface{} `json:"__APOLLO_STATE__"`
			} `json:"pageProps"`
		} `json:"props"`
		Query struct {
			WorkID string `json:"workId"`
		} `json:"query"`
	}

	if err := json.Unmarshal([]byte(script), &kakuyomuNextData); err != nil {
		return "", nil, metadata, fmt.Errorf("kakuyomu:__NEXT_DATA__ parse error %w", err)
	}

	if kakuyomuNextData.Props.PageProps.ApolloState == nil {
		return "", nil, metadata, errors.New("kakuyomu:__APOLLO_STATE__ not found")
	}

	storyId := kakuyomuNextData.Query.WorkID

	authorWork, ok := kakuyomuNextData.Props.PageProps.ApolloState["Work:"+storyId].(map[string]interface{})
	if !ok {
		return "", nil, metadata, errors.New("kakuyomu:work(author) not found")
	}

	title, ok := authorWork["title"].(string)
	if !ok {
		return "", nil, metadata, errors.New("kakuyomu:title not found or broken type")
	}

	updated, err := getTimFromObj(authorWork["lastEpisodePublishedAt"])
	if err != nil {
		return "", nil, metadata, fmt.Errorf("kakuyomu:lastEpisodePublishedAt not found or broken %w", err)
	}

	desc, ok := authorWork["introduction"].(string)
	if !ok {
		return "", nil, metadata, errors.New("kakuyomu:introduction not found or broken type")
	}

	desc = trimDescription(desc)

	authorRef, ok := authorWork["author"].(map[string]interface{})
	if !ok {
		return "", nil, metadata, errors.New("kakuyomu:authorRef not found or broken type")
	}
	authorRefId, ok := authorRef["__ref"].(string)
	if !ok {
		return "", nil, metadata, errors.New("kakuyomu:authorRefId not found or broken type")
	}
	authorAccount, ok := kakuyomuNextData.Props.PageProps.ApolloState[authorRefId].(map[string]interface{})
	if !ok {
		return "", nil, metadata, errors.New("kakuyomu:UserAccount(author) not found or broken type")
	}
	author, ok := authorAccount["activityName"].(string)
	if !ok {
		return "", nil, metadata, errors.New("kakuyomu:ActivityName not found or broken type")
	}

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: fmt.Sprintf("https://kakuyomu.jp/works/%s", storyId)},
		Description: desc,
		Author:      &feeds.Author{Name: author},
		Updated:     updated,
	}

	for _, v := range kakuyomuNextData.Props.PageProps.ApolloState {
		if v == nil {
			continue
		}
		it, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if it["__typename"] == "Episode" {

			id, ok := it["id"].(string)
			if !ok {
				continue
			}

			title, ok := it["title"].(string)
			if !ok {
				continue
			}

			publishedAt, err := getTimFromObj(it["publishedAt"])
			if err != nil {
				continue
			}

			uri := fmt.Sprintf("https://kakuyomu.jp/works/%s/episodes/%s", storyId, id)

			feed.Items = append(feed.Items, &feeds.Item{
				Title:   title,
				Link:    &feeds.Link{Href: uri},
				Id:      id,
				Created: publishedAt,
			})

		}
	}

	if len(feed.Items) == 0 {
		return "", nil, metadata, fmt.Errorf("kakuyomu:no episode entry")
	}

	sort.Slice(feed.Items, func(i, j int) bool {
		return feed.Items[i].Created.Before(feed.Items[j].Created)
	})

	return "kakuyomu_works" + storyId, feed, metadata, nil
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
