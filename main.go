package main

import (
	termboxutil "../termboxutil"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"regexp"
	"strings"
)

var titleRegex *regexp.Regexp // the regex used to cleanup filenames
var cache string              // cache directory $HOME/.movietin

const (
	omdbUrl = "http://www.omdbapi.com/?%s=%s&plot=full"
)

/*
holds the response from omdbapi */
type Movie struct {
	Title, Year, Genre, Director, Actors, Plot, ImdbID string
}

type SearchResult struct {
	Search []Movie
}

func cacheLookup(fname string) (movie Movie, found error) {
	file, openerr := os.OpenFile(strings.Trim(cache + fname, " "), os.O_RDONLY, 0660)
	if openerr != nil {
		return movie, errors.New("No file in cache " + cache + fname)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&movie); err != nil {
		return movie, errors.New("Failed decoding gob in cache")
	}

	return
}

func cacheSave(fname string, movie Movie) error {
	file, openerr := os.OpenFile(strings.Trim(cache + fname, " "), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0660)
	if openerr != nil {
		return errors.New("Couldnt open cache for saving")
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(movie); err != nil {
		return errors.New("Couldnt save cache")
	}

	return nil
}

func omdbLookup(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	data, readerr := ioutil.ReadAll(resp.Body)
	if readerr != nil {
		panic(readerr)
	}
	resp.Body.Close()

	return data
}

func main() {

	cwd, _ := os.Getwd()
	dir, _ := os.Open(cwd)
	media, err := dir.Readdir(-1)
	if err != nil {
		panic(err)
	}

	u, _ := user.Current()
	cache = u.HomeDir + "/.movietin/"

	// TODO make this cleaner and replace periods etc
	// regex to cleanup filenames from extensions and misc data
	titleRegex, _ = regexp.Compile("[\\w\\d\\s]+")

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetCursor(-1, -1)

	filenames := make([]string, len(media))

	for i, file := range media {
		prettyName := titleRegex.FindString(file.Name())
		cachedImdb, cacherr := cacheLookup(prettyName)

		if cacherr != nil {
			filenames[i] = prettyName
		} else {
			filenames[i] = cachedImdb.Title + "\t" + cachedImdb.Year
		}

	}

	screen := termboxutil.Screen{}

	mainWindow := screen.NewWindow(
		termbox.ColorWhite,
		termbox.ColorDefault,
		termbox.ColorGreen,
		termbox.ColorBlack)
	mainWindow.Scrollable(true)

	searchWindow := screen.NewWindow(
		termbox.ColorWhite,
		termbox.ColorDefault,
		termbox.ColorGreen,
		termbox.ColorBlack)
	searchWindow.Scrollable(true)

	err = mainWindow.Draw(filenames)
	screen.Focus(&mainWindow)

	if err != nil {
		panic(err)
	}
	termbox.Flush()

	var searchResult []Movie
	mainWindow.CatchEvent = func(event termbox.Event) {
		if event.Ch == 'j' || event.Key == termbox.KeyArrowDown {
			mainWindow.NextRow()
		} else if event.Ch == 'k' || event.Key == termbox.KeyArrowUp {
			mainWindow.PrevRow()
		} else if event.Key == termbox.KeyEnter {

			// do a search for titleRegex match of the filename
			curRow, _ := mainWindow.CurrentRow()
			searchData := omdbLookup(fmt.Sprintf(
				omdbUrl,
				"s",
				url.QueryEscape(titleRegex.FindString(curRow.Text))))

			var sr SearchResult
			err = json.Unmarshal(searchData, &sr)
			if err != nil {
				panic(err)
			}
			searchResult = sr.Search

			titles := make([]string, len(searchResult))

			for i, movie := range searchResult {
				titles[i] = movie.Title
			}

			searchWindow.Draw(titles)
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
		} else if event.Key == termbox.KeyEnter {
			currentRow, index := searchWindow.CurrentRow()

			// do movie lookup by id
			lookupData := omdbLookup(fmt.Sprintf(
				omdbUrl,
				"i",
				searchResult[index].ImdbID))

			var movie Movie
			err = json.Unmarshal(lookupData, &movie)
			if err != nil {
				panic(err)
			}

			err = cacheSave(titleRegex.FindString(currentRow.Text), movie)
			if err != nil {
				panic(err)
			}
			screen.Focus(&mainWindow)
			mainWindow.Redraw()
			termbox.Flush()
			return
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
