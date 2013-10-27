package main

import (
	"encoding/gob"
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
// TODO rename this
var cache string // full cache $HOME/cacheBase

const (
	cacheBase = "/.tinbox/"
	omdbUrl   = "http://www.omdbapi.com/?t=%s&plot=full"
)

/*
holds the response from omdbapi */
type Movie struct {
	Title, Year, Genre, Director, Actors, Plot string
}

/*
TODO make filename parsing more flexible */
func fileFmt(name string) string {
	idx := strings.Index(name, " (")
	name = name[:idx]

	filename := cache + name

	cacheFile, err := os.OpenFile(filename, os.O_RDONLY, 0660)

	if os.IsNotExist(err) {
		resp, err := http.Get(fmt.Sprintf(omdbUrl, url.QueryEscape(name)))
		if err != nil {
			return name
		}

		data, readerr := ioutil.ReadAll(resp.Body)
		if readerr != nil {
			panic(readerr)
		}
		resp.Body.Close()

		result := Movie{}
		err = json.Unmarshal(data, &result)

		if err != nil {
			panic(err)
		}

		outfile, ioerr := os.Create(filename)
		if ioerr != nil {
			panic(ioerr)
		}

		genc := gob.NewEncoder(outfile)
		genc.Encode(result)
		outfile.Close()

		cacheFile, err = os.OpenFile(filename, os.O_RDONLY, 0660)

		if err != nil {
			panic(err)
		}
	}

	gdec := gob.NewDecoder(cacheFile)
	defer cacheFile.Close()

	movie := Movie{}
	gdec.Decode(&movie)

	return movie.Title
}

func draw(media []os.FileInfo) {

	xpos := 0
	ypos := 0

	_, maxy := term.Size()
	if selected < 0 {
		selected = 0
		return
	} else if selected == maxy || selected == len(media) {
		selected = selected - 1 // reset selected (TODO dont do this)
		return
	}

	for i, file := range media {

		if file.IsDir() {
			continue
		} else if ypos == maxy {
			break // break since we are at the end of the display
		}

		for _, ch := range fileFmt(file.Name()) {
			if i == selected {
				term.SetCell(xpos, ypos, rune(ch), term.ColorGreen, term.ColorBlack)
			} else {
				term.SetCell(xpos, ypos, rune(ch), term.ColorWhite, term.ColorBlack)
			}
			xpos += 1
		}

		ypos += 1
		xpos = 0
	}

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

	draw(media)

	for {
		// redraw any changes
		term.Flush()

		// wait for new events
		e := term.PollEvent()

		// handle resize
		if e.Type == term.EventResize {
			term.Clear(term.ColorWhite, term.ColorBlack)
			draw(media)
			continue
		}

		// handle error
		if e.Type == term.EventError {
			continue
		}

		// handle keys
		if e.Key == term.KeyArrowUp || e.Ch == 'k' {
			selected -= 1
		} else if e.Key == term.KeyArrowDown || e.Ch == 'j' {
			selected += 1
		}

		draw(media)
	}
}
