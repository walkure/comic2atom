package siteloader

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

func GetFeed(ctx context.Context, target string) (string, *feeds.Feed, error) {
	uri, err := url.Parse(target)
	if err != nil {
		return "", nil, err
	}

	if strings.HasPrefix(target, "https://storia.takeshobo.co.jp/manga/") {
		return storiaFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://gammaplus.takeshobo.co.jp/manga/") {
		return gammaPlusFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://comic-meteor.jp/") {
		return meteorFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://www.comicride.jp/book/") {
		return rideFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://www.comic-valkyrie.com/") {
		return valkyrieFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://ncode.syosetu.com/") {
		return narouFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://kakuyomu.jp/works/") {
		return kakuyomuFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://comic-fuz.com/manga/") {
		return fuzFeed(ctx, uri)
	}

	if strings.HasPrefix(target, "https://comic-walker.com/detail/") {
		return comicwalkerFeed(ctx, uri)
	}

	return "", nil, fmt.Errorf("%s not supported site", target)
}

func escapePath(path string) string {
	var sb strings.Builder
	for _, r := range path {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r == '_') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func trimDescription(desc string) string {

	normalizedDesc := strings.NewReplacer(
		"\r\n", "\n",
		"\r", "\n",
	).Replace(desc)

	var sb strings.Builder
	for _, line := range strings.Split(normalizedDesc, "\n") {
		sb.WriteString(strings.TrimSpace(line))
	}

	return sb.String()
}

func resolveRelativeURI(baseUri *url.URL, relative string) (string, error) {
	relativeUri, err := baseUri.Parse(relative)
	if err != nil {
		return "", fmt.Errorf("relative:%w", err)
	}

	return relativeUri.String(), nil
}

func fetchDocument(ctx context.Context, target *url.URL) (*goquery.Document, error) {
	// HTTP Get
	req, err := http.NewRequestWithContext(ctx, "GET", target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot generate request:%w", err)
	}
	req.Header.Set("User-Agent", "Saitama")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP/GET error:%w", err)
	}
	defer res.Body.Close()

	// Read
	bytesRead, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read error:%w", err)
	}

	// detect charset
	detector := chardet.NewTextDetector()
	deetctResult, err := detector.DetectBest(bytesRead)
	if err != nil {
		return nil, fmt.Errorf("charset detect error:%w", err)
	}

	// convert charset
	bytesReader := bytes.NewReader(bytesRead)
	reader, err := charset.NewReaderLabel(deetctResult.Charset, bytesReader)
	if err != nil {
		return nil, fmt.Errorf("charset convert error:%w", err)
	}

	// create document
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("cannot create goquery document:%w", err)
	}

	return doc, nil
}

func generateHashedHex(id string) string {
	return fmt.Sprintf("%x", (md5.Sum([]byte(id))))
}
