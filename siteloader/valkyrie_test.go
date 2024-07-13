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

func TestValkyrie(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/valkyrie_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, err := valkyrieFeed(context.Background(), testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "valkyrie_path_test", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テスト", feed.Author.Name)
	assert.Equal(t, "テストてすとストーリー", feed.Description)

	testcases := []struct {
		path  string
		thumb string
		title string
	}{
		{
			path:  "https://test/03",
			thumb: "img/thm/ep_003.jpg",
			title: "テスト3",
		},
		{
			path:  "https://test/02",
			thumb: "img/thm/ep_002.jpg",
			title: "テスト2",
		},
		{
			path:  "https://test/01",
			thumb: "img/thm/ep_001.jpg",
			title: "テスト1",
		},
	}

	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			absimg, _ := resolveRelativeURI(testUrl, tt.thumb)
			assert.Equal(t, generateHashedHex(tt.path), feed.Items[index].Id)
			assert.Equal(t, tt.path, feed.Items[index].Link.Href)
			assert.Equal(t, absimg, feed.Items[index].Enclosure.Url)
			assert.Equal(t, tt.title, feed.Items[index].Title)
		})
	}
	assert.Panics(t, func() { _ = feed.Items[3].Title })
}
