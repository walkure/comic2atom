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

func TestAlphapolisMangaOfficial(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/alphapolis_mo_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, _, err := alphapolisMOFeed(context.Background(), testUrl)
	assert.Nil(t, err)
	assert.Equal(t, "alphapolis_path_test", fname)
	assert.Equal(t, "テスト作品タイトル", feed.Title)
	assert.Equal(t, "これは検証用のあらすじです。", feed.Description)
	assert.Equal(t, "テスト著者/漫画 | テスト著者/原作", feed.Author.Name)

	// 無料エピソードのみが取得されること（第3話は有料なので除外される）
	assert.Equal(t, 2, len(feed.Items))

	testcases := []struct {
		path  string
		thumb string
		title string
	}{
		{
			path:  "/manga/official/456/1001",
			thumb: "https://example.com/thumb1.jpg",
			title: "第1話",
		},
		{
			path:  "/manga/official/456/1002",
			thumb: "https://example.com/thumb2.jpg",
			title: "第2話",
		},
	}

	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			absPath, _ := resolveRelativeURI(testUrl, tt.path)

			assert.Equal(t, generateHashedHex(absPath), feed.Items[index].Id)
			assert.Equal(t, absPath, feed.Items[index].Link.Href)
			assert.Equal(t, tt.thumb, feed.Items[index].Enclosure.Url)
			assert.Equal(t, tt.title, feed.Items[index].Title)
		})
	}
}
