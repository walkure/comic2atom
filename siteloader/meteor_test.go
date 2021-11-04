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

	fname, doc, err := meteorFeed(testUrl)

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

	fname, feed, err := meteorFeed(testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "meteor_path_test", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テストてすとストーリー", feed.Description)
	assert.Equal(t, "テスト名", feed.Author.Name)

	assert.Equal(t, "eeec23eabaa19622c2dff466251aa48a", feed.Items[0].Id)
	assert.Equal(t, "https://example.com/test/5", feed.Items[0].Link.Href)
	assert.Equal(t, "テスト5", feed.Items[0].Title)

	assert.Equal(t, "00610eae8358ad655e54e5275009fe21", feed.Items[1].Id)
	assert.Equal(t, "https://example.com/test/3", feed.Items[1].Link.Href)
	assert.Equal(t, "テスト3", feed.Items[1].Title)

	assert.Equal(t, "f2ce4bdb1b346b98f28027cbb71e4432", feed.Items[2].Id)
	assert.Equal(t, "https://example.com/test/2", feed.Items[2].Link.Href)
	assert.Equal(t, "テスト2", feed.Items[2].Title)

	assert.Equal(t, "8471aefb4d405fe08651deebad1479de", feed.Items[3].Id)
	assert.Equal(t, "https://example.com/test/1", feed.Items[3].Link.Href)
	assert.Equal(t, "テスト1", feed.Items[3].Title)

	assert.Panics(t, func() { _ = feed.Items[4].Title })
}
