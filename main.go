package main

import (
	termutil "../termboxutil"
	"encoding/json"
	"fmt"
	term "github.com/nsf/termbox-go"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
)

var selected int // index of selected title
var cache string // full cache $HOME/cacheBase

const (
	cacheBase = "/.tinbox/"
	omdbUrl   = "http://www.omdbapi.com/?%s=%s&plot=%s"
)

/*
holds the response from omdbapi */
type Movie struct {
	Title, Year, Genre, Director, Actors, Plot, ImdbID string
}

type SearchResult struct {
	Search []Movie
}

func omdbSearch(title string) SearchResult {
	resp, err := http.Get(fmt.Sprintf(omdbUrl, "s", url.QueryEscape(title), "short"))
	if err != nil {
		panic(err)
	}

	data, readerr := ioutil.ReadAll(resp.Body)
	if readerr != nil {
		panic(readerr)
	}
	resp.Body.Close()

	var result SearchResult
	err = json.Unmarshal(data, &result)

	if err != nil {
		panic(err)
	}

	return result
}

func main() {

	cwd, _ := os.Getwd()
	dir, _ := os.Open(cwd)
	media, err := dir.Readdir(-1)
	if err != nil {
		panic(err)
	}

	u, _ := user.Current()
	cache = u.HomeDir + cacheBase

	if err := term.Init(); err != nil {
		panic(err)
	}
	defer term.Close()
	term.SetCursor(-1, -1)

	filenames := make([]string, len(media))

	for i, file := range media {
		filenames[i] = file.Name()
	}

	screen := termutil.NewScreen(term.ColorWhite, term.ColorDefault, term.ColorGreen, term.ColorBlack)
	screen.Scrollable(true)

	searchScreen := termutil.NewScreen(term.ColorWhite, term.ColorDefault, term.ColorGreen, term.ColorBlack)
	searchScreen.Scrollable(true)

	err = screen.Draw(filenames)
	if err != nil {
		panic(err)
	}

	selected = 0

	for {
		for {

			// redraw any changes
			term.Flush()

			// wait for new events
			e := term.PollEvent()

			// handle resize
			if e.Type == term.EventResize {
				err = screen.Redraw()
				if err != nil {
					panic(err)
				}
				continue
			}

			// handle error
			if e.Type == term.EventError {
				panic(e.Err)
				continue
			}

			// handle keys
			if e.Key == term.KeyArrowUp || e.Ch == 'k' {
				screen.PrevRow()
			} else if e.Key == term.KeyArrowDown || e.Ch == 'j' {
				screen.NextRow()
			} else if e.Ch == 'i' {
				break
			}

			err = screen.Redraw()

			if err != nil {
				panic(err)
			}
		}

		search := screen.CurrentRow().Text
		idx := strings.LastIndex(search, " (")

		search = search[:idx]

		searchResult := omdbSearch(search)

		titles := make([]string, len(searchResult.Search))

		for i, movie := range searchResult.Search {
			titles[i] = movie.Title
		}

		searchScreen.Draw(titles)

		// loop for searches
		for {

			term.Flush()

			e := term.PollEvent()

			// handle resize
			if e.Type == term.EventResize {
				err = searchScreen.Redraw()
				if err != nil {
					panic(err)
				}
				continue
			}

			// handle error
			if e.Type == term.EventError {
				panic(e.Err)
			}

			// handle keys
			if e.Key == term.KeyArrowUp || e.Ch == 'k' {
				searchScreen.PrevRow()
			} else if e.Key == term.KeyArrowDown || e.Ch == 'j' {
				searchScreen.NextRow()
			}

			err = searchScreen.Redraw()

			if err != nil {
				panic(err)
			}
		}
	}
}
