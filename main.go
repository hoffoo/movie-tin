package main

import (
	termboxutil "../termboxutil"
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
)

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
	idx := strings.Index(title, " (")
	title = title[:idx]
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

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetCursor(-1, -1)

	filenames := make([]string, len(media))

	for i, file := range media {
		filenames[i] = file.Name()
	}

	screen := termboxutil.Screen{}

	mainWindow := screen.NewWindow(termbox.ColorWhite, termbox.ColorDefault, termbox.ColorGreen, termbox.ColorBlack)
	mainWindow.Scrollable(true)

	searchWindow := screen.NewWindow(termbox.ColorWhite, termbox.ColorDefault, termbox.ColorGreen, termbox.ColorBlack)
	searchWindow.Scrollable(true)

	err = mainWindow.Draw(filenames)
	screen.Focus(&mainWindow)

	if err != nil {
		panic(err)
	}
	termbox.Flush()

	mainWindow.CatchEvent = func(event termbox.Event) {
		if event.Ch == 'j' || event.Key == termbox.KeyArrowDown {
			mainWindow.NextRow()
		} else if event.Ch == 'k' || event.Key == termbox.KeyArrowUp {
			mainWindow.PrevRow()
		} else if event.Ch == 'i' {
			searchResult := omdbSearch(mainWindow.CurrentRow().Text)
			searchData := make([]string, len(searchResult.Search))

			for i, movieResult := range searchResult.Search {
				searchData[i] = movieResult.Title
			}
			searchWindow.Draw(searchData)
			screen.Focus(&searchWindow)
			termbox.Flush()
			return
		}

		mainWindow.Redraw()
		termbox.Flush()
	}

	searchWindow.CatchEvent = func(event termbox.Event) {
		if event.Ch == 'j' || event.Key == termbox.KeyArrowDown {
			searchWindow.NextRow()
		} else if event.Ch == 'k' || event.Key == termbox.KeyArrowUp {
			searchWindow.PrevRow()
		} else if event.Ch == 'q' || event.Key == termbox.KeyEsc {
			screen.Focus(&mainWindow)
			mainWindow.Redraw()
			termbox.Flush()
			return
		}

		searchWindow.Redraw()
		termbox.Flush()
	}

	screen.Loop()
}
