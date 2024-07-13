package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/walkure/comic2atom/siteloader"
)

var listener = flag.String("listener", ":8080", "listen address and port")

func main() {
	flag.Parse()

	// default router NOT remains double slashes.
	r := mux.NewRouter().SkipClean(true)
	r.PathPrefix("/entry/").HandlerFunc(handleEntry)

	fmt.Printf("server starting at %s\n", *listener)
	fmt.Printf("server shutting down:%+v", http.ListenAndServe(*listener, r))
}

func handleEntry(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	rawuri := strings.TrimPrefix(r.URL.Path, "/entry/")

	_, feed, err := siteloader.GetFeed(r.Context(), rawuri)
	if err != nil {
		fmt.Printf("GetFeed error:%+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feedXml, err := feed.ToAtom()
	if err != nil {
		fmt.Printf("ToAtom error:%+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprint(w, feedXml)

}
