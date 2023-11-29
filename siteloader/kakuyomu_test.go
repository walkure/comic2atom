package siteloader

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKakuyomu(t *testing.T) {
	testsv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/kakuyomu_test.html")
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	}))
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/works/987654321")

	fname, feed, err := kakuyomuFeed(testUrl)
	assert.Nil(t, err)

	assert.Equal(t, "kakuyomu_works987654321", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テスト著者", feed.Author.Name)
	assert.Equal(t, "テストてすとストーリー", feed.Description)

	wantTime := parseTestDate(t, "2023-09-30 10:25:58 (UTC)")
	assert.True(t, wantTime.Equal(feed.Updated),
		"(updated)want %v,got %v", wantTime, feed.Updated)

	testcases := []struct {
		path    string
		id      string
		title   string
		created string
	}{
		{
			path:    "/works/987654321/episodes/1111111",
			id:      "1111111",
			title:   "第1話",
			created: "2020-10-09 06:13:22 (UTC)",
		},
		{
			path:    "/works/987654321/episodes/2222222",
			id:      "2222222",
			title:   "第2話",
			created: "2020-10-24 23:41:18 (UTC)",
		},
		{
			path:    "/works/987654321/episodes/3333333",
			id:      "3333333",
			title:   "第3話",
			created: "2021-01-16 22:00:06 (UTC)",
		},
		{
			path:    "/works/987654321/episodes/4444444",
			id:      "4444444",
			title:   "第4話",
			created: "2021-01-17 22:00:02 (UTC)",
		},
		{
			path:    "/works/987654321/episodes/5555555",
			id:      "5555555",
			title:   "第5話",
			created: "2023-05-13 07:17:24 (UTC)",
		},
	}

	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			abspath, _ := resolveRelativeURI(testUrl, tt.path)

			assert.Equal(t, tt.id, feed.Items[index].Id)
			assert.Equal(t, abspath, feed.Items[index].Link.Href)
			assert.Equal(t, tt.title, feed.Items[index].Title)

			wantTime := parseTestDate(t, tt.created)
			assert.True(t, wantTime.Equal(feed.Items[index].Created),
				"(created)want %v,got %v", wantTime, feed.Items[index].Created)
		})
	}

}
