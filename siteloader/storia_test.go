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

func TestStoriaErr(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "example")
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	fname, doc, err := storiaFeed(testUrl)

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

	fname, feed, err := storiaFeed(testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "storia_path_test", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テストてすとストーリー", feed.Description)
	assert.Equal(t, "テスト名", feed.Author.Name)

	abspath, _ := resolveRelativeURI(testUrl, "./_files/5/")
	absimg, _ := resolveRelativeURI(testUrl, "./_img/5.jpg")
	assert.Equal(t, generateHashedHex(abspath), feed.Items[0].Id)
	assert.Equal(t, abspath, feed.Items[0].Link.Href)
	assert.Equal(t, absimg, feed.Items[0].Enclosure.Url)
	assert.Equal(t, "テスト5", feed.Items[0].Title)

	abspath, _ = resolveRelativeURI(testUrl, "./est#comics")
	assert.Equal(t, generateHashedHex(abspath), feed.Items[1].Id)
	assert.Equal(t, abspath, feed.Items[1].Link.Href)
	assert.Equal(t, "テスト2", feed.Items[1].Title)

	abspath, _ = resolveRelativeURI(testUrl, "./_files/01/")
	absimg, _ = resolveRelativeURI(testUrl, "./_img/1.jpg")
	assert.Equal(t, generateHashedHex(abspath), feed.Items[2].Id)
	assert.Equal(t, abspath, feed.Items[2].Link.Href)
	assert.Equal(t, absimg, feed.Items[2].Enclosure.Url)
	assert.Equal(t, "テスト1", feed.Items[2].Title)

	assert.Panics(t, func() { _ = feed.Items[3].Title })
}
