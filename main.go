package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"golang.org/x/net/html"
)

var webpagesVisited = make(map[string]*Webpage, 0)

func main() {
	_, err := InitializeDB()
	if err != nil {
		log.Fatal(err)
	}

	wew, err := NewWebpage("https://404notboring.com/articles/boids")
	// wew, err := NewWebpage("https://en.wikipedia.org/wiki/Mud")
	if err != nil {
		fmt.Println(err)
	}

	err = wew.PopulateWebpageInfo()
	if err != nil {
		fmt.Println(err)
	}

	for {
		var input int
		fmt.Println("On site:", wew.URL)
		for _, v := range wew.LinkedFrom {
			fmt.Println("Linked from:", v.URL)
		}
		for i, v := range wew.LinkedTo {
			fmt.Println(i, ")", v.URL)
		}
		fmt.Println("Enter a number: ")
		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println("Invalid input, try again")
			continue
		}
		if input >= 0 && input < len(wew.LinkedTo) {
			wew = wew.LinkedTo[input]
			if wew.HTML == "" {
				wew.PopulateWebpageInfo()
			}
		} else {
			fmt.Println("Number must be between 0 and", len(wew.LinkedTo)-1)
		}

	}

}

func InitializeDB() (*sql.DB, error) {
	// db stuff
	db, err := sql.Open("sqlite3", "./mydb.sqlite")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// apply migrations
	if err := goose.Up(db, "db/migrations"); err != nil {
		return nil, err
	}
	return db, nil
}

type Webpage struct {
	URL        string
	Lang       string
	HTML       string
	LinkedTo   []*Webpage
	LinkedFrom []*Webpage
}

func NewWebpage(URL string) (*Webpage, error) {
	URL, err := ResolveHrefToURL(URL, "")
	if err != nil {
		return nil, err
	}

	if webpagesVisited[URL] != nil {
		return nil, errors.New("webpage already exists")
	} else {
		webpage := &Webpage{
			URL: URL,
		}

		webpagesVisited[URL] = webpage

		return webpage, nil
	}
}

func (webpage *Webpage) PopulateWebpageInfo() error {
	resp, err := http.Get(webpage.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return z.Err()
		case html.StartTagToken:
			tn, _ := z.TagName()

			for {
				attribute, value, moreAttr := z.TagAttr()
				if !moreAttr {
					break
				}
				if string(tn) == "a" && string(attribute) == "href" {
					href := string(value)
					URL, err := ResolveHrefToURL(href, webpage.URL)
					if err != nil {
						break
					}
					webpage.LinkMeTo(URL)
				}
			}
		}
	}
}

func (webpage *Webpage) LinkMeTo(URL string) {
	isNewURL := webpagesVisited[URL] == nil
	if isNewURL {
		wp, err := NewWebpage(URL)
		if err != nil {
			return
		}
		webpagesVisited[URL] = wp
		wp.LinkMeFrom(webpage)
		webpage.LinkedTo = append(webpage.LinkedTo, wp)
	} else {
		wp := webpagesVisited[URL]
		alreadyLinkedTo := slices.Contains(webpage.LinkedTo, wp)

		if !alreadyLinkedTo {
			webpage.LinkedTo = append(webpage.LinkedTo, wp)
			wp.LinkMeFrom(webpage)
		}
	}

}

func (webpage *Webpage) LinkMeFrom(from *Webpage) {
	if !slices.Contains(webpage.LinkedFrom, from) {
		webpage.LinkedFrom = append(webpage.LinkedFrom, from)
	}
}

func (webpage *Webpage) Print() {
	for _, link := range webpage.LinkedTo {
		fmt.Println(link.URL)
	}
}

func ResolveHrefToURL(href string, currentURL string) (string, error) {
	u, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	if currentURL == "" {
		return u.String(), nil
	}

	base, err := url.Parse(currentURL)
	if err != nil {
		return "", err
	}

	if base.Scheme == "" {
		return "", errors.New("no scheme on currentURL")
	}

	joined := base.ResolveReference(u).String()
	return joined, nil
}
