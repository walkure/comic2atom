package siteloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/feeds"
)

func ganganonlineFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {
	doc, metadata, err := fetchDocument(ctx, target)
	if err != nil {
		return "", nil, metadata, fmt.Errorf("ganganonline:FetchErr:%w", err)
	}

	script := doc.Find("script#__NEXT_DATA__").Text()
	if script == "" {
		return "", nil, metadata, errors.New("ganganonline:__NEXT_DATA__ not found")
	}

	var ganganonlineNextData struct {
		Props struct {
			PageProps struct {
				Data struct {
					Default struct {
						Chapters []struct {
							ID               int    `json:"id"`
							Status           int    `json:"status,omitempty"`
							ThumbnailURL     string `json:"thumbnailUrl"`
							MainText         string `json:"mainText"`
							SubText          string `json:"subText,omitempty"`
							AppLaunchURL     string `json:"appLaunchUrl,omitempty"`
							PublishingPeriod string `json:"publishingPeriod,omitempty"`
						} `json:"chapters"`
						RelatedTitleLinks []any  `json:"relatedTitleLinks"`
						TitleName         string `json:"titleName"`
						ImageURL          string `json:"imageUrl"`
						Author            string `json:"author"`
						Description       string `json:"description"`
						TitleID           int    `json:"titleId"`
					} `json:"default"`
				} `json:"data"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(script), &ganganonlineNextData); err != nil {
		return "", nil, metadata, fmt.Errorf("ganganonline:__NEXT_DATA__ parse error %w", err)
	}

	defaultData := ganganonlineNextData.Props.PageProps.Data.Default

	feed := &feeds.Feed{
		Title:       defaultData.TitleName,
		Link:        &feeds.Link{Href: target.String()},
		Description: trimDescription(defaultData.Description),
		Author:      &feeds.Author{Name: defaultData.Author},
		Created:     time.Now(),
	}

	for _, chapter := range defaultData.Chapters {
		if chapter.Status != 0 {
			continue
		}
		uri := fmt.Sprintf("https://www.ganganonline.com/title/%d/chapter/%d", defaultData.TitleID, chapter.ID)
		feed.Items = append(feed.Items, &feeds.Item{
			Title: chapter.MainText,
			Link:  &feeds.Link{Href: uri},
			Id:    generateHashedHex(uri),
		})
	}

	if len(feed.Items) == 0 {
		return "", nil, metadata, fmt.Errorf("ganganonline:no episode entry")
	}

	return fmt.Sprintf("ganganonline_%d", defaultData.TitleID), feed, metadata, nil
}
