package main

import (
	"bufio"
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
	list           = flag.String("list", "", "targets url(s) list")
	atomPathPrefix = flag.String("atom", "", "atom file save path prefix")
)

func init() {
	flag.Parse()
}

func main() {
	if (*targets == "" && *list == "") || *atomPathPrefix == "" {
		log.Fatal("requires target,list and atom arguments.")
	}

	var targetUris []string

	if *targets != "" {
		if strings.Contains(*targets, ",") {
			targetUris = append(targetUris, strings.Split(*targets, ",")...)

		} else {
			targetUris = append(targetUris, *targets)
		}
	}

	if *list != "" {
		loaded, err := loadList(*list)
		if err != nil {
			fmt.Printf("cannot load file(%s):%v", *list, err)
		}
		targetUris = append(targetUris, loaded...)
	}

	if len(targetUris) == 0 {
		fmt.Printf("no target found from args(%s) nor list(%s)", *targets, *list)
	}

	errored := false
	for _, target := range targetUris {
		err := processTarget(target, *atomPathPrefix)
		if err != nil {
			fmt.Printf("Error:%v\n", err)
			errored = true
		}
	}

	if errored {
		os.Exit(255)
	}
}

func loadList(listPath string) ([]string, error) {
	fp, err := os.Open(listPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	var list []string
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" && line[0] != '#' {
			list = append(list, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return list, nil

}

func processTarget(targetUri, pathPrefix string) error {

	fmt.Printf("Fetch %s ", targetUri)

	fname, feed, err := siteloader.GetFeed(targetUri)

	if err != nil {
		return err
	}

	atomData, err := feed.ToAtom()
	if err != nil {
		return err
	}

	atomPath := filepath.Join(pathPrefix, "/", fname+".atom")

	file, err := os.OpenFile(atomPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	file.WriteString(atomData)
	file.Close()

	fmt.Printf("-> %s\n", atomPath)
	return nil
}
