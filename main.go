package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/walkure/comic2atom/siteloader"
)

var (
	targets        = flag.String("targets", "", "check target uri(s)")
	atomPathPrefix = flag.String("atom", "", "atom file save path prefix")
)

func init() {
	flag.Parse()
}

func main() {
	if *targets == "" || *atomPathPrefix == "" {
		log.Fatal("requires target and atom arguments.")
	}

	if strings.Contains(*targets, ",") {
		for _, target := range strings.Split(*targets, ",") {
			processTarget(target, *atomPathPrefix)
		}

	} else {
		processTarget(*targets, *atomPathPrefix)
	}
}

func processTarget(targetUri, pathPrefix string) {

	fname, feed, err := siteloader.GetFeed(targetUri)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fetch %s ", targetUri)

	atomData, err := feed.ToAtom()
	if err != nil {
		log.Fatal(err)
	}

	atomPath := filepath.Join(pathPrefix, "/", fname+".atom")

	file, err := os.OpenFile(atomPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}

	file.WriteString(atomData)
	file.Close()

	fmt.Printf("-> %s\n", atomPath)
}
