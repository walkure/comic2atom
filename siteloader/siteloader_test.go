package siteloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapePath(t *testing.T) {
	assert.Equal(t, "sa1Tama", escapePath("sa1Tama"))
	assert.Equal(t, "sa1ama", escapePath("sa1/ama"))
	assert.Equal(t, "sa1_ama", escapePath("sa1_ama"))
}

func TestTrimDescription(t *testing.T) {
	assert.Equal(t, "sa1Tama", trimDescription("sa1Tama"))
	assert.Equal(t, "sa1Tama", trimDescription(" sa1 \nTama"))
	assert.Equal(t, "sa1Tama", trimDescription("    sa1 \r Tama"))
}

func TestResolveRelativeURI(t *testing.T) {

	base, _ := url.Parse("https://www.example.com/")

	result, err := resolveRelativeURI(base, "saitama")
	assert.Nil(t, err)
	assert.Equal(t, "https://www.example.com/saitama", result)

	result, err = resolveRelativeURI(base, "sait/ama/")
	assert.Nil(t, err)
	assert.Equal(t, "https://www.example.com/sait/ama/", result)

	result, err = resolveRelativeURI(base, "sait/ama/")
	assert.Nil(t, err)
	assert.Equal(t, "https://www.example.com/sait/ama/", result)
}

func TestFetchDocumentPrimitive(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "example")
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	doc, _, err := fetchDocument(context.Background(), testUrl)

	if err != nil {
		t.Errorf("%v", err)
	}

	assert.Equal(t, "example", doc.Text())
}

func TestFetchDocumentFailure(t *testing.T) {
	// https://pod.hatenablog.com/entry/2021/03/10/081909
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("for test -- close connection")
		conn.Close()
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	_, _, err := fetchDocument(context.Background(), testUrl)

	assert.Error(t, err)

	// https://github.com/stretchr/testify/issues/1066
	if !errors.Is(err, io.EOF) {
		t.Errorf("EOF is expected but return error is %[1]T, %+[1]v", err)
	}
}

func TestGetFeed(t *testing.T) {
	fname, feed, _, err := GetFeed(context.Background(), "https://www.example.com/")
	assert.Equal(t, "", fname)
	assert.Nil(t, feed)
	assert.NotNil(t, err)

	fname, feed, _, err = GetFeed(context.Background(), "hoge")
	assert.Equal(t, "", fname)
	assert.Nil(t, feed)
	assert.NotNil(t, err)
}

func TestGenerateHashedHex(t *testing.T) {
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", generateHashedHex(""))
	assert.Equal(t, "1a79a4d60de6718e8e5b326e338ae533", generateHashedHex("example"))
}

func TestFetchDocumentConditionalRequest(t *testing.T) {
	var exampleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == "d41d8cd98f00b204e9800998ecf8427e" {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		if r.Header.Get("If-Modified-Since") == "Wed, 21 Oct 2015 07:28:00 GMT" {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		fmt.Fprintf(w, "example")
	})

	testsv := httptest.NewServer(exampleHandler)
	defer testsv.Close()

	testUrl, _ := url.Parse(testsv.URL)

	ctx := SetIfNoneMatch(context.Background(), "d41d8cd98f00b204e9800998ecf8427e")
	_, _, err := fetchDocument(ctx, testUrl)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, ErrNotModified))

	ctx = SetIfModifiedSince(context.Background(), "Wed, 21 Oct 2015 07:28:00 GMT")
	_, _, err = fetchDocument(ctx, testUrl)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, ErrNotModified))
}
