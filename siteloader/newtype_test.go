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

func TestNewtype(t *testing.T) {

	var fileProvider = func(filePath string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Cannot load test file:%v", err)
			}
			defer f.Close()
			io.Copy(w, f)
		}
	}
	//
	mux := http.NewServeMux()
	mux.HandleFunc("/contents/test/more/1/Dsc", fileProvider("./testdata/newtype_test.json"))
	mux.HandleFunc("/contents/test/", fileProvider("./testdata/newtype_test.html"))
	testsv := httptest.NewServer(mux)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/contents/test/32")

	fname, feed, _, err := newtypeFeed(context.Background(), testUrl)
	assert.Nil(t, err)
	assert.Equal(t, "newtype_contentstest", fname)
	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testsv.URL+"/contents/test/", feed.Link.Href)
	assert.Equal(t, "テスト内容", feed.Description)
	assert.Equal(t, "テスト作者", feed.Author.Name)

	testcases := []struct {
		path    string
		thumb   string
		title   string
		created string
	}{
		{
			path:    "/contents/test/20/",
			thumb:   "/rsz/C1/img/comic/test/thumbnail_02.jpg/w200/",
			title:   "第２話",
			created: "2024-08-02 00:00:00 (JST)",
		},
		{
			path:    "/contents/test/10/",
			thumb:   "/rsz/C1/img/comic/test/thumbnail_01.jpg/w200/",
			title:   "第１話",
			created: "2024-07-19 00:00:00 (JST)",
		},
	}
	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			abspath, _ := resolveRelativeURI(testUrl, tt.path)
			absimg, _ := resolveRelativeURI(testUrl, tt.thumb)
			assert.Equal(t, generateHashedHex(abspath+tt.title), feed.Items[index].Id)
			wantTime := parseTestDate(t, tt.created)
			assert.Equal(t, abspath, feed.Items[index].Link.Href)
			assert.Equal(t, absimg, feed.Items[index].Enclosure.Url)
			assert.Equal(t, tt.title, feed.Items[index].Title)
			assert.True(t, wantTime.Equal(feed.Items[index].Created),
				"(created)want %v,got %v", wantTime, feed.Items[index].Created)
		})
	}

	assert.Panics(t, func() { _ = feed.Items[3].Title })

}
