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

	assert.Equal(t, "3f0cbc9dce91157284cc751df185e7b5", feed.Items[0].Id)
	assert.Equal(t, "https://www.example.com/episode4", feed.Items[0].Link.Href)
	assert.Equal(t, "テスト4", feed.Items[0].Title)

	assert.Equal(t, "38bf86656cb26e37f01562466369422f", feed.Items[1].Id)
	assert.Equal(t, "https://www.example.com/episode3", feed.Items[1].Link.Href)
	assert.Equal(t, "テスト3", feed.Items[1].Title)

	assert.Equal(t, "429f57c494cdc837398855fdc0e8d6d4", feed.Items[2].Id)
	assert.Equal(t, "https://www.example.com/episode2", feed.Items[2].Link.Href)
	assert.Equal(t, "テスト2", feed.Items[2].Title)

	assert.Equal(t, "60221f61ce824b808cb5b598b1e10aa4", feed.Items[3].Id)
	assert.Equal(t, "https://www.example.com/episode1", feed.Items[3].Link.Href)
	assert.Equal(t, "テスト1", feed.Items[3].Title)

	assert.Panics(t, func() { _ = feed.Items[4].Title })
}
