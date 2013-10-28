package main

import (
	termutil "../termboxutil"
	term "github.com/nsf/termbox-go"
	"os"
	"os/user"
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

type Search struct {
	Search []Movie
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

	screen := termutil.NewScreen(term.ColorWhite, term.ColorBlack)

	err = screen.Draw(filenames)
	if err != nil {
		panic(err)
	}

	selected = 0

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

			if selected-1 >= 0 {
				selected -= 1
			} else {
				continue
			}

			screen.UnmarkRow(selected+1)
			screen.MarkRow(selected, term.ColorGreen, term.ColorBlack)
		} else if e.Key == term.KeyArrowDown || e.Ch == 'j' {

			if selected+1 < len(screen.Rows) {
				selected += 1
			} else {
				continue
			}

			screen.UnmarkRow(selected-1)
			screen.MarkRow(selected, term.ColorGreen, term.ColorBlack)
		}

		err = screen.Redraw()

		if err != nil {
			panic(err)
		}
	}
}
