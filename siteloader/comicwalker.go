package siteloader

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

func comicwalkerFeed(target *url.URL) (string, *feeds.Feed, error) {
	doc, err := fetchDocument(target)
	if err != nil {
		return "", nil, fmt.Errorf("comicwalker:FetchErr:%w", err)
	}

	script := doc.Find("script#__NEXT_DATA__").Text()
	if script == "" {
		return "", nil, errors.New("comicwalker:__NEXT_DATA__ not found")
	}

	var walkerNextData struct {
		Props struct {
			PageProps struct {
				WorkCode        string `json:"workCode"`
				DehydratedState struct {
					Queries []map[string]interface{} `json:"queries"`
				} `json:"dehydratedState"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(script), &walkerNextData); err != nil {
		return "", nil, fmt.Errorf("comicwalker:__NEXT_DATA__ parse error %w", err)
	}

	if len(walkerNextData.Props.PageProps.DehydratedState.Queries) == 0 {
		return "", nil, errors.New("comicwalker:Queries not found")
	}

	detailJSON, err := getComicDetailJSON(walkerNextData.Props.PageProps.DehydratedState.Queries)
	if err != nil {
		return "", nil, fmt.Errorf("comicwalker:DetailErr:%w", err)
	}

	var comicDetail struct {
		Data struct {
			Work struct {
				Title   string `json:"title"`
				Summary string `json:"summary"`
				Authors []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Role string `json:"role"`
				} `json:"authors"`
			} `json:"work"`
			FirstEpisodes struct {
				Total  int `json:"total"`
				Result []struct {
					ID             string        `json:"id"`
					Code           string        `json:"code"`
					Title          string        `json:"title"`
					SubTitle       string        `json:"subTitle"`
					Thumbnail      string        `json:"thumbnail"`
					UpdateDate     time.Time     `json:"updateDate"`
					DeliveryPeriod time.Time     `json:"deliveryPeriod"`
					IsNew          bool          `json:"isNew"`
					HasRead        bool          `json:"hasRead"`
					Stores         []interface{} `json:"stores"`
					ServiceID      string        `json:"serviceId"`
					Internal       struct {
						EpisodeNo   int    `json:"episodeNo"`
						PageCount   int    `json:"pageCount"`
						Episodetype string `json:"episodetype"`
					} `json:"internal"`
					Type     string `json:"type"`
					IsActive bool   `json:"isActive"`
				} `json:"result"`
			} `json:"firstEpisodes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(detailJSON, &comicDetail); err != nil {
		return "", nil, fmt.Errorf("comicwalker:Detailed JSON parse error %w", err)
	}

	authors := make([]string, 0, len(comicDetail.Data.Work.Authors))
	for _, a := range comicDetail.Data.Work.Authors {
		authors = append(authors, fmt.Sprintf("%s(%s)", a.Name, a.Role))
	}

	feed := &feeds.Feed{
		Title:       comicDetail.Data.Work.Title,
		Link:        &feeds.Link{Href: fmt.Sprintf("https://comic-walker.com/detail/%s", walkerNextData.Props.PageProps.WorkCode)},
		Description: trimDescription(comicDetail.Data.Work.Summary),
		Author:      &feeds.Author{Name: strings.Join(authors, ", ")},
	}

	for _, ep := range comicDetail.Data.FirstEpisodes.Result {
		if !ep.IsActive {
			continue
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       ep.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("https://comic-walker.com/detail/%s/episodes/%s", walkerNextData.Props.PageProps.WorkCode, ep.Code)},
			Id:          ep.ID,
			Created:     ep.UpdateDate,
			Description: ep.SubTitle,
		})
		feed.Updated = ep.UpdateDate
	}

	return "comicwalker_" + walkerNextData.Props.PageProps.WorkCode, feed, nil
}

func getComicDetailJSON(data []map[string]interface{}) ([]byte, error) {
	for _, query := range data {
		for k, v := range query {
			if k == "queryKey" {
				queryArray, ok := v.([]interface{})
				if !ok {
					return nil, errors.New("queryKey is not array")
				}
				for _, queryParams := range queryArray {
					queryPath, ok := queryParams.(string)
					if !ok {
						continue
					}
					fmt.Printf("%+v\n", queryPath)
					if queryPath == "/api/contents/details/work" {
						return json.Marshal(query["state"])
					}
				}
			}
		}
	}
	return nil, errors.New("comic detail not found")
}
