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

	myURL, err := url.Parse("https://404notboring.com/articles/boids")
	if err != nil {
		fmt.Println(err)
	}

	wew, err := NewWebpage(myURL)
	if err != nil {
		fmt.Println(err)
	}

	err = wew.PopulateWebpageInfo()
	if err != nil {
		fmt.Println(err)
	}

	for {
		var input int
		fmt.Println("On site:", wew.URL, "\""+wew.Title+"\"")
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
	URL        *url.URL
	Lang       string
	HTML       string
	Title      string
	LinkedTo   []*Webpage
	LinkedFrom []*Webpage
}

func NewWebpage(URL *url.URL) (*Webpage, error) {
	if webpagesVisited[URL.String()] != nil {
		return nil, errors.New("webpage already exists")
	} else {
		webpage := &Webpage{
			URL: URL,
		}

		webpagesVisited[URL.String()] = webpage

		return webpage, nil
	}
}

func (webpage *Webpage) PopulateWebpageInfo() error {
	resp, err := http.Get(webpage.URL.String())
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
			tagName, _ := z.TagName()

			switch string(tagName) {
			case "title":
				if z.Next() == html.TextToken {
					title := string(z.Text())
					webpage.Title = title
				}
			case "a":
				for {
					attribute, value, moreAttr := z.TagAttr()
					if string(attribute) == "href" {
						href := string(value)
						URL, err := ResolveHrefToURL(href, webpage.URL.String())
						if err != nil {
							break
						}
						webpage.LinkMeTo(URL)
					}
					if !moreAttr {
						break
					}
				}
			}
		}
	}
}

func (webpage *Webpage) LinkMeTo(URL *url.URL) {
	isNewURL := webpagesVisited[URL.String()] == nil
	if isNewURL {
		wp, err := NewWebpage(URL)
		if err != nil {
			return
		}
		webpagesVisited[URL.String()] = wp
		wp.LinkMeFrom(webpage)
		webpage.LinkedTo = append(webpage.LinkedTo, wp)
	} else {
		wp := webpagesVisited[URL.String()]
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

func ResolveHrefToURL(href string, currentURL string) (*url.URL, error) {
	u, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	if currentURL == "" {
		return u, nil
	}

	base, err := url.Parse(currentURL)
	if err != nil {
		return nil, err
	}

	if base.Scheme == "" {
		return nil, errors.New("no scheme on currentURL")
	}

	joined := base.ResolveReference(u)
	return joined, nil
}
