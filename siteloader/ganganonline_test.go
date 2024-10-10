package siteloader

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGanganOnline(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/ganganonline_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, _, err := ganganonlineFeed(context.Background(), testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "ganganonline_12345", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テストてすとストーリー", feed.Description)
	assert.Equal(t, "テスト名", feed.Author.Name)

	testcases := []struct {
		path  string
		title string
		desc  string
	}{
		{
			path:  "https://www.ganganonline.com/title/12345/chapter/333",
			title: "第3話",
		},
		{
			path:  "https://www.ganganonline.com/title/12345/chapter/222",
			title: "第2話",
		},
		{
			path:  "https://www.ganganonline.com/title/12345/chapter/111",
			title: "第1話",
		},
	}
	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(t, tt.path, feed.Items[index].Link.Href)
			assert.Equal(t, tt.title, feed.Items[index].Title)
			assert.Equal(t, generateHashedHex(tt.path), feed.Items[index].Id)
		})
	}

	assert.Panics(t, func() { _ = feed.Items[4].Title })
}
