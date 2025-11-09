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

func GetFeed(ctx context.Context, target string) (string, *feeds.Feed, HttpMetadata, error) {
	uri, err := url.Parse(target)
	if err != nil {
		return "", nil, HttpMetadata{}, err
	}

	if strings.HasPrefix(target, "https://kirapo.jp/") {
		return meteorFeed(ctx, uri)
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

	if strings.HasPrefix(target, "https://www.ganganonline.com/title/") {
		return ganganonlineFeed(ctx, uri)
	}

	return "", nil, HttpMetadata{}, fmt.Errorf("%s not supported site", target)
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

const ifNoneMatchKey = ifNoneMatchType("If-None-Match")

type ifNoneMatchType string

const ifModifiedSinceKey = isModifiedSinceType("If-Modified-Since")

type isModifiedSinceType string

var ErrNotModified = fmt.Errorf("content not modified")

func getIfNoneMatch(ctx context.Context) (string, bool) {
	if ifNoneMatch, ok := ctx.Value(ifNoneMatchKey).(string); ok {
		return ifNoneMatch, true
	}
	return "", false
}

func getIsModifiedSince(ctx context.Context) (string, bool) {
	if isModifiedSince, ok := ctx.Value(ifModifiedSinceKey).(string); ok {
		return isModifiedSince, true
	}
	return "", false
}

func SetIfNoneMatch(ctx context.Context, ifNoneMatch string) context.Context {
	if ifNoneMatch == "" {
		return ctx
	}
	return context.WithValue(ctx, ifNoneMatchKey, ifNoneMatch)
}

func SetIfModifiedSince(ctx context.Context, ifModifiedSince string) context.Context {
	if ifModifiedSince == "" {
		return ctx
	}
	return context.WithValue(ctx, ifModifiedSinceKey, ifModifiedSince)
}

type HttpMetadata struct {
	ETag         string
	LastModified string
}

func getHttpBody(ctx context.Context, target *url.URL) ([]byte, HttpMetadata, error) {

	// HTTP Get
	req, err := http.NewRequestWithContext(ctx, "GET", target.String(), nil)
	if err != nil {
		return nil, HttpMetadata{}, fmt.Errorf("cannot generate request:%w", err)
	}
	req.Header.Set("User-Agent", "Saitama")

	//set if-none-match and if-modified-since
	if ifNoneMatch, ok := getIfNoneMatch(ctx); ok {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
	if ifModifiedSince, ok := getIsModifiedSince(ctx); ok {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, HttpMetadata{}, fmt.Errorf("HTTP/GET error:%w", err)
	}
	defer res.Body.Close()

	// Not Modified
	if res.StatusCode == http.StatusNotModified {
		return nil, HttpMetadata{}, ErrNotModified
	}

	// Read
	bytesRead, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, HttpMetadata{}, fmt.Errorf("read error:%w", err)
	}

	return bytesRead, HttpMetadata{
		ETag:         res.Header.Get("ETag"),
		LastModified: res.Header.Get("Last-Modified"),
	}, nil
}

func fetchDocument(ctx context.Context, target *url.URL) (*goquery.Document, HttpMetadata, error) {

	bytesRead, metadata, err := getHttpBody(ctx, target)
	if err != nil {
		return nil, metadata, err
	}

	// detect charset
	detector := chardet.NewTextDetector()
	deetctResult, err := detector.DetectBest(bytesRead)
	if err != nil {
		return nil, metadata, fmt.Errorf("charset detect error:%w", err)
	}

	// convert charset
	bytesReader := bytes.NewReader(bytesRead)
	reader, err := charset.NewReaderLabel(deetctResult.Charset, bytesReader)
	if err != nil {
		return nil, metadata, fmt.Errorf("charset convert error:%w", err)
	}

	// create document
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, metadata, fmt.Errorf("cannot create goquery document:%w", err)
	}

	return doc, metadata, nil
}

func generateHashedHex(id string) string {
	return fmt.Sprintf("%x", (md5.Sum([]byte(id))))
}
