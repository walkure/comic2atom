package siteloader

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNarou(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fn := "./testdata/narou_test_p1.html"

		if r.URL.Query().Get("p") != "" {
			fn = "./testdata/narou_test_p2.html"
		}

		f, err := os.Open(fn)
		if err != nil {
			t.Fatalf("Cannot load test file:%v", err)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL + "/path_t/est")

	fname, feed, err := narouFeed(testUrl)
	assert.Nil(t, err)
	assert.Nil(t, err)

	assert.Equal(t, "narou_path_test", fname)

	assert.Equal(t, "テストタイトル", feed.Title)
	assert.Equal(t, testUrl.String(), feed.Link.Href)
	assert.Equal(t, "テスト著者", feed.Author.Name)
	assert.Equal(t, "テストてすとストーリー", feed.Description)

	feedWantUpdated := parseTestDate(t, "2022-05-28 11:12:00 (JST)")
	assert.True(t, feedWantUpdated.Equal(feed.Updated),
		"(feed updated)want %v,got %v", feedWantUpdated, feed.Updated)

	testcases := []struct {
		path    string
		title   string
		created string
		updated string
	}{
		{
			path:    "/novelid/1/",
			title:   "チャプター1/サブタイトル1",
			created: "2022-05-26 18:00:00 (JST)",
			updated: "2022-05-27 18:41:00 (JST)",
		},
		{
			path:    "/novelid/2/",
			title:   "チャプター1/サブタイトル2",
			created: "2022-05-26 19:00:00 (JST)",
			updated: "",
		},
		{
			path:    "/novelid/3/",
			title:   "チャプター2/サブタイトル3",
			created: "2022-05-27 16:00:00 (JST)",
			updated: "2022-05-28 11:12:00 (JST)",
		},
		{
			path:    "/novelid/4/",
			title:   "チャプター2/サブタイトル4",
			created: "2022-05-27 20:00:00 (JST)",
			updated: "",
		},
	}

	for index, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			abspath, _ := resolveRelativeURI(testUrl, tt.path)

			assert.Equal(t, generateHashedHex(abspath), feed.Items[index].Id)
			assert.Equal(t, abspath, feed.Items[index].Link.Href)
			assert.Equal(t, tt.title, feed.Items[index].Title)

			wantTime := parseTestDate(t, tt.created)
			assert.True(t, wantTime.Equal(feed.Items[index].Created),
				"(created)want %v,got %v", wantTime, feed.Items[index].Created)

			wantTime = parseTestDate(t, tt.updated)
			assert.True(t, wantTime.Equal(feed.Items[index].Updated),
				"(updated)want %v,got %v", wantTime, feed.Items[index].Updated)
		})
	}
	assert.Panics(t, func() { _ = feed.Items[4].Title })
}

func Test_parseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    string
		wantErr bool
	}{
		{
			name:    "SUCCESS(with update)",
			arg:     "2022/05/27 18:41 改稿",
			want:    "2022-05-27 18:41:00 (JST)",
			wantErr: false,
		},
		{
			name:    "SUCCESS(with update 2)",
			arg:     "2022/05/27 19:00 改",
			want:    "2022-05-27 19:00:00 (JST)",
			wantErr: false,
		},
		{
			name:    "SUCCESS(simple)",
			arg:     "2022/05/27 19:00",
			want:    "2022-05-27 19:00:00 (JST)",
			wantErr: false,
		},
		{
			name:    "FAILURE(too short)",
			arg:     "2022/05/2719:00",
			wantErr: true,
		},
		{
			name:    "FAILURE(too long)",
			arg:     "2022/05/27   19:00",
			wantErr: true,
		},
		{
			name:    "FAILURE(invalid)",
			arg:     "hogehogehogehogehogehoge",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimestamp(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				wantTime := parseTestDate(t, tt.want)
				if !got.Equal(wantTime) {
					t.Errorf("parseTimestamp() = %v, want %v", got, wantTime)
				}
			}
		})
	}
}

func parseTestDate(t *testing.T, date string) time.Time {
	t.Helper()
	if date == "" {
		return time.Time{}
	}
	tm, err := time.Parse("2006-01-02 15:04:05 (MST)", date)
	if err != nil {
		t.Fatalf("toDate: %v", err)
	}

	return tm
}
