package siteloader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"google.golang.org/protobuf/proto"
)

func fuzFeed(target *url.URL) (string, *feeds.Feed, error) {
	idx := strings.LastIndex(target.Path, "/")
	idStr := target.Path[idx+1:]
	freeOnly := target.Query().Has("freeOnly")

	tq := target.Query()
	tq.Del("freeOnly")
	target.RawQuery = tq.Encode()

	if idStr == "" {
		return "", nil, errors.New("fuz:invalid URI")
	}

	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return "", nil, fmt.Errorf("fuz:invalid id: %w", err)
	}

	mangaId := uint32(id64)

	mdReq := &MangaDetailRequest{
		MangaId: &mangaId,
		DeviceInfo: &DeviceInfo{
			DeviceType: DeviceInfo_BROWSER,
		},
	}

	req, err := proto.Marshal(mdReq)
	if err != nil {
		return "", nil, fmt.Errorf("fuz:failure to marshal request: %w", err)
	}

	res, err := http.Post("https://api.comic-fuz.com/v1/manga_detail", "application/protobuf", bytes.NewReader(req))
	if err != nil {
		return "", nil, fmt.Errorf("fuz:failure to post request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("fuz:server failed: %d(%s)", res.StatusCode, http.StatusText(res.StatusCode))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil, fmt.Errorf("fuz:failure to read request: %w", err)
	}

	data := &MangaDetailResponse{}

	if err = proto.Unmarshal(body, data); err != nil {
		return "", nil, fmt.Errorf("fuz:failure to unmarshal response: %w", err)
	}

	var authors []string
	for _, as := range data.Authorship {
		for _, ath := range as.Author {
			authors = append(authors, ath.AuthorName)
		}
	}

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", nil, fmt.Errorf("fuz:failure to load Asia/Tokyo timezone: %w", err)
	}

	latestUpdate, err := time.ParseInLocation("2006/01/02", data.Manga.LatestUpdatedDate, loc)
	if err != nil {
		return "", nil, fmt.Errorf("fuz:failure to parse LatestUpdatedDate[%s]: %w", data.Manga.LatestUpdatedDate, err)
	}

	feed := &feeds.Feed{
		Title:       data.Manga.MangaName,
		Link:        &feeds.Link{Href: target.String()},
		Description: data.Manga.LongDescription,
		Author:      &feeds.Author{Name: strings.Join(authors, "/")},
		Created:     latestUpdate,
	}

	for _, cg := range data.Chapters {
		for _, c := range cg.Chapters {
			if freeOnly && c.PointConsumption.GetAmount() != 0 {
				continue
			}

			title := c.ChapterMainName
			if c.ChapterSubName != "" {
				title = c.ChapterMainName + "/" + c.ChapterSubName
			}

			at, err := time.ParseInLocation("2006/01/02", c.UpdatedDate, loc)
			if err != nil {
				continue
			}
			href := fmt.Sprintf("https://comic-fuz.com/manga/viewer/%d", c.ChapterId)
			feed.Items = append(feed.Items, &feeds.Item{
				Title:   title,
				Updated: at,
				Link:    &feeds.Link{Href: href},
				Id:      generateHashedHex(href),
			})
		}
	}

	freeOnlyPrefix := ""
	if freeOnly {
		freeOnlyPrefix = "_freeOnly"
	}

	return "fuz_" + escapePath(target.Path) + freeOnlyPrefix, feed, nil
}
