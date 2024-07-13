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

func TestStoriaErr(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "example")
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	fname, doc, _, err := storiaFeed(context.Background(), testUrl)

	assert.Error(t, err)
	assert.Equal(t, "", fname)
	assert.Nil(t, doc)
}

func TestStoria(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/storia_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, _, err := storiaFeed(context.Background(), testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "storia_path_test", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テストてすとストーリー", feed.Description)
	assert.Equal(t, "テスト名", feed.Author.Name)

	testcases := []struct {
		path  string
		thumb string
		title string
	}{
		{
			path:  "./_files/5/",
			thumb: "./_img/5.jpg",
			title: "テスト5",
		},
		{
			path:  "./est#comics",
			thumb: "../../path_t/_files/5/",
			title: "テスト2",
		},
		{
			path:  "./_files/01/",
			thumb: "./_img/1.jpg",
			title: "テスト1",
		},
	}
	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			abspath, _ := resolveRelativeURI(testUrl, tt.path)
			absimg, _ := resolveRelativeURI(testUrl, tt.thumb)
			assert.Equal(t, generateHashedHex(abspath), feed.Items[index].Id)
			assert.Equal(t, abspath, feed.Items[index].Link.Href)
			assert.Equal(t, absimg, feed.Items[index].Enclosure.Url)
			assert.Equal(t, tt.title, feed.Items[index].Title)
		})
	}

	assert.Panics(t, func() { _ = feed.Items[3].Title })
}
