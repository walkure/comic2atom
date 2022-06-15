package siteloader

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

func GetFeed(target string) (string, *feeds.Feed, error) {
	uri, err := url.Parse(target)
	if err != nil {
		return "", nil, err
	}

	if strings.HasPrefix(target, "https://storia.takeshobo.co.jp/manga/") {
		return storiaFeed(uri)
	}

	if strings.HasPrefix(target, "https://gammaplus.takeshobo.co.jp/manga/") {
		return gammaPlusFeed(uri)
	}

	if strings.HasPrefix(target, "https://comic-meteor.jp/") {
		return meteorFeed(uri)
	}

	if strings.HasPrefix(target, "https://www.comicride.jp/book/") {
		return rideFeed(uri)
	}

	if strings.HasPrefix(target, "https://www.comic-valkyrie.com/") {
		return valkyrieFeed(uri)
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

func fetchDocument(target *url.URL) (*goquery.Document, error) {

	// HTTP Get
	res, err := http.Get(target.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP/GET error:%w", err)
	}
	defer res.Body.Close()

	// Read
	bytesRead, err := ioutil.ReadAll(res.Body)
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
