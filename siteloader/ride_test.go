package siteloader

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRideErr(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "example")
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	fname, doc, err := rideFeed(testUrl)

	assert.Error(t, err)
	assert.Equal(t, "", fname)
	assert.Nil(t, doc)
}

func TestRide(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/ride_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, err := rideFeed(testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "ride_path_test", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テスト名", feed.Author.Name)
	assert.Equal(t, "テスト内容", feed.Description)

	testcases := []struct {
		path  string
		hash  string
		title string
	}{
		{
			title: "テスト4",
			hash:  "3f0cbc9dce91157284cc751df185e7b5",
			path:  "https://www.example.com/episode4",
		},
		{
			title: "テスト3",
			hash:  "38bf86656cb26e37f01562466369422f",
			path:  "https://www.example.com/episode3",
		},
		{
			title: "テスト2",
			hash:  "429f57c494cdc837398855fdc0e8d6d4",
			path:  "https://www.example.com/episode2",
		},
		{
			title: "テスト1",
			hash:  "60221f61ce824b808cb5b598b1e10aa4",
			path:  "https://www.example.com/episode1",
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
