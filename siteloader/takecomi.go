package siteloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

func takecomiFeed(ctx context.Context, target *url.URL) (string, *feeds.Feed, HttpMetadata, error) {
	idx := strings.LastIndex(target.Path, "/")
	idStr := target.Path[idx+1:]
	if idStr == "" {
		return "", nil, HttpMetadata{}, errors.New("takecomi:invalid URI")
	}

	// try to get total eposodes
	var seriesSimpleData struct {
		Series struct {
			Summary struct {
				ID          string   `json:"id"`
				Name        string   `json:"name"`
				Description string   `json:"description"`
				PublishDate UnixTime `json:"publishDate"`
				UpdatedOn   UnixTime `json:"updatedOn"`
				Status      string   `json:"status"`
				Author      []struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
					Role string `json:"role"`
				} `json:"author"`
				IsCompleted bool `json:"isCompleted"`
				NumEpisodes int  `json:"numEpisodes"`
				IsUp        bool `json:"isUp"`
			} `json:"summary"`
			Episodes []struct {
				ID            string      `json:"id"`
				IndexID       int         `json:"indexId"`
				Title         string      `json:"title"`
				URL           interface{} `json:"url"`
				DatePublished UnixTime    `json:"datePublished"`
			} `json:"episodes"`
		} `json:"series"`
	}
	seriesSimpleUrl := "https://takecomic.jp/api/episodes?seriesHash=" + idStr
	seriesSimpleResp, err := http.Get(seriesSimpleUrl)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("takecomi:failure to fetch(simple) %q :%w", seriesSimpleUrl, err)
	}
	defer seriesSimpleResp.Body.Close()
	err = json.NewDecoder(seriesSimpleResp.Body).Decode(&seriesSimpleData)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("takecomi:failure to decode(simple) %q :%w", seriesSimpleUrl, err)
	}

	episodeFrom := max(1, seriesSimpleData.Series.Summary.NumEpisodes-5)
	description, err := dequoteTalecomiDetails(seriesSimpleData.Series.Summary.Description)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("takecomi:failure to read description %q :%w", seriesSimpleUrl, err)
	}

	// get eposode details
	seriesDetailUrl := fmt.Sprintf("https://takecomic.jp/api/episodes?episodeFrom=%d&episodeTo=%d&seriesHash=%s", episodeFrom, seriesSimpleData.Series.Summary.NumEpisodes, idStr)

	seriesResp, err := http.Get(seriesDetailUrl)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("takecomi:failure to fetch %q :%w", seriesDetailUrl, err)
	}
	defer seriesResp.Body.Close()

	var seriesDetailData struct {
		Series struct {
			Summary struct {
				ID          string   `json:"id"`
				Name        string   `json:"name"`
				Description string   `json:"description"`
				PublishDate UnixTime `json:"publishDate"`
				UpdatedOn   UnixTime `json:"updatedOn"`
				Status      string   `json:"status"`
				Author      []struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
					Role string `json:"role"`
				} `json:"author"`
				IsCompleted bool `json:"isCompleted"`
				NumEpisodes int  `json:"numEpisodes"`
				IsUp        bool `json:"isUp"`
			} `json:"summary"`
			Episodes []struct {
				ID            string      `json:"id"`
				IndexID       int         `json:"indexId"`
				Title         string      `json:"title"`
				URL           interface{} `json:"url"`
				DatePublished UnixTime    `json:"datePublished"`
			} `json:"episodes"`
		} `json:"series"`
	}

	err = json.NewDecoder(seriesResp.Body).Decode(&seriesDetailData)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("takecomi:failure to decode %q :%w", seriesDetailUrl, err)
	}

	accessUrl := fmt.Sprintf("https://takecomic.jp/api/series/access?episodeFrom=%d&episodeTo=%d&seriesHash=%s", episodeFrom, seriesSimpleData.Series.Summary.NumEpisodes, idStr)
	accessResp, err := http.Get(accessUrl)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("takecomi:failure to fetch %q :%w", accessUrl, err)
	}
	defer accessResp.Body.Close()

	var accessData struct {
		SeriesAccess struct {
			SeriesID        string `json:"seriesId"`
			EpisodeAccesses []struct {
				EpisodeID      string `json:"episodeId"`
				HasAccess      bool   `json:"hasAccess"`
				IsCampaign     bool   `json:"isCampaign"`
				AccessType     string `json:"accessType"`
				Read           bool   `json:"read"`
				IsDisableBonus bool   `json:"isDisableBonus,omitempty"`
				Price          int    `json:"price,omitempty"`
			} `json:"episodeAccesses"`
		} `json:"seriesAccess"`
	}

	err = json.NewDecoder(accessResp.Body).Decode(&accessData)
	if err != nil {
		return "", nil, HttpMetadata{}, fmt.Errorf("takecomi:failure to decode %q :%w", accessUrl, err)
	}

	accessMap := make(map[string]bool, len(accessData.SeriesAccess.EpisodeAccesses))
	for _, ac := range accessData.SeriesAccess.EpisodeAccesses {
		accessMap[ac.EpisodeID] = ac.HasAccess
	}

	authors := make([]string, 0, len(seriesDetailData.Series.Summary.Author))
	for _, ac := range seriesDetailData.Series.Summary.Author {
		authors = append(authors, fmt.Sprintf("%s(%s)", ac.Name, ac.Role))
	}

	feed := &feeds.Feed{
		Title:       seriesDetailData.Series.Summary.Name,
		Link:        &feeds.Link{Href: target.String()},
		Description: description,
		Author:      &feeds.Author{Name: strings.Join(authors, "/")},
		Created:     time.Time(seriesDetailData.Series.Summary.PublishDate),
		Updated:     time.Time(seriesDetailData.Series.Summary.UpdatedOn),
	}

	for _, ep := range seriesDetailData.Series.Episodes {
		if !accessMap[ep.ID] {
			continue
		}
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:   ep.Title,
				Updated: time.Time(ep.DatePublished),
				Id:      ep.ID,
				Link: &feeds.Link{
					Href: "https://takecomic.jp/episodes/" + ep.ID,
				},
			})
	}

	return "takecomi_" + escapePath(target.Path), feed, HttpMetadata{}, nil
}

func dequoteTalecomiDetails(details string) (string, error) {
	var detailsData []struct {
		Type     string `json:"type"`
		Children []struct {
			Text string `json:"text"`
		} `json:"children"`
	}

	err := json.Unmarshal([]byte(details), &detailsData)
	if err != nil {
		return "", fmt.Errorf("failure to unmarshal %q: %w", details, err)
	}

	sb := strings.Builder{}
	for _, d := range detailsData {
		for _, dd := range d.Children {
			sb.WriteString(dd.Text)
		}
	}

	return sb.String(), nil
}
