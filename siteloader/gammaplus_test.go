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

func TestGammaPlusErr(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "example")
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	fname, doc, err := gammaPlusFeed(testUrl)

	assert.Error(t, err)
	assert.Equal(t, "", fname)
	assert.Nil(t, doc)
}

func TestGammaPlus(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/gammaplus_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, err := gammaPlusFeed(testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "gammaplus_path_test", fname)

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
			path:  "./_files/5/",
			title: "テスト5",
			desc:  "テスト5D",
		},
		{
			path:  "./_files/4/",
			title: "テスト4",
			desc:  "テスト4D",
		},
		{
			path:  "./est",
			title: "テスト2",
			desc:  "",
		},
		{
			path:  "./_files/01/",
			title: "テスト1",
			desc:  "テスト1D",
		},
	}
	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			abspath, _ := resolveRelativeURI(testUrl, tt.path)
			assert.Equal(t, generateHashedHex(abspath), feed.Items[index].Id)
			assert.Equal(t, abspath, feed.Items[index].Link.Href)
			assert.Equal(t, tt.title, feed.Items[index].Title)
			assert.Equal(t, tt.desc, feed.Items[index].Description)
		})
	}

	assert.Panics(t, func() { _ = feed.Items[4].Title })
}
