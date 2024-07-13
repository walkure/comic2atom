package siteloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTrimmedAuthor(t *testing.T) {
	assert.Equal(t, "ほげ", getTrimmedAuthor(" 著者　 　：ほげ "))
}

func TestMeteorErr(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "example")
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	fname, doc, err := meteorFeed(context.Background(), testUrl)

	assert.Error(t, err)
	assert.Equal(t, "", fname)
	assert.Nil(t, doc)
}

func TestMeteor(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/meteor_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, err := meteorFeed(context.Background(), testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "meteor_path_test", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テストてすとストーリー", feed.Description)
	assert.Equal(t, "テスト名", feed.Author.Name)

	testcases := []struct {
		path  string
		hash  string
		title string
	}{
		{
			path:  "https://example.com/test/5",
			hash:  "eeec23eabaa19622c2dff466251aa48a",
			title: "テスト5",
		},
		{
			path:  "https://example.com/test/3",
			hash:  "00610eae8358ad655e54e5275009fe21",
			title: "テスト3",
		},
		{
			path:  "https://example.com/test/2",
			hash:  "f2ce4bdb1b346b98f28027cbb71e4432",
			title: "テスト2",
		},
		{
			path:  "https://example.com/test/1",
			hash:  "8471aefb4d405fe08651deebad1479de",
			title: "テスト1",
		},
	}

	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(t, tt.hash, feed.Items[index].Id)
			assert.Equal(t, tt.path, feed.Items[index].Link.Href)
			assert.Equal(t, tt.title, feed.Items[index].Title)
		})
	}

	assert.Panics(t, func() { _ = feed.Items[4].Title })
}
