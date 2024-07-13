package siteloader

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComicWalker(t *testing.T) {
	testsv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/comicwalker_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	}))
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/detail/KC_WCODE_SAMPLE")
	fname, feed, _, err := comicwalkerFeed(context.Background(), testUrl)
	assert.Nil(t, err)

	assert.Equal(t, "comicwalker_KC_WCODE_SAMPLE", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, "https://comic-walker.com/detail/KC_WCODE_SAMPLE", feed.Link.Href)
	assert.Equal(t, "テスト原作(原作), テスト著者(著者)", feed.Author.Name)
	assert.Equal(t, "テストてすとストーリー", feed.Description)

	wantTime, _ := time.Parse(time.RFC3339, "2024-06-30T02:00:00Z")
	assert.True(t, wantTime.Equal(feed.Updated),
		"(updated)want %v,got %v", wantTime, feed.Updated)

	testcases := []struct {
		path    string
		id      string
		title   string
		created string
	}{
		{
			path:    "https://comic-walker.com/detail/KC_WCODE_SAMPLE/episodes/STORY_CODE_1",
			id:      "018fc1af-7bc4-785c-a307-321f877e6dc9",
			title:   "第一話",
			created: "2024-04-30T02:00:00Z",
		},
		{
			path:    "https://comic-walker.com/detail/KC_WCODE_SAMPLE/episodes/STORY_CODE_3",
			id:      "01905d8f-105f-75e9-8be5-5c44c8aeacb2",
			title:   "第三話",
			created: "2024-06-30T02:00:00Z",
		},
	}

	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(t, tt.id, feed.Items[index].Id)
			assert.Equal(t, tt.path, feed.Items[index].Link.Href)
			assert.Equal(t, tt.title, feed.Items[index].Title)

			wantTime, _ := time.Parse(time.RFC3339, tt.created)
			assert.True(t, wantTime.Equal(feed.Items[index].Created),
				"(created)want %v,got %v", wantTime, feed.Items[index].Created)
		})
	}

}
