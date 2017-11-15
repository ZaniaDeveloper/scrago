package scrago

import (
	"os"
	"encoding/json"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type Record map[string]interface{}

type ParseFunc func(*goquery.Document, chan<- Record)

type Spider struct {
	Name      string
	StartURLs []string
	Parse     ParseFunc
}

func get_doc(doc *goquery.Document, err error) *goquery.Document {
	if err != nil {
		panic(err)
	}

	return doc
}

func get_doc_from_filename(urlString string) *goquery.Document {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(u.Path, os.O_RDONLY, 0444)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	return get_doc(goquery.NewDocumentFromReader(f))
}

func get_doc_from_url(urlString string) *goquery.Document {
	return get_doc(goquery.NewDocument(urlString))
}

func run_scrape(urlString string, c chan<- Record, parser ParseFunc) {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	var doc_getter func(string) *goquery.Document

	if u.Scheme == "file" {
		doc_getter = get_doc_from_filename
	} else {
		doc_getter = get_doc_from_url
	}

	parser(doc_getter(urlString), c)
}

func RunSpider(spider Spider) {
	records := make([]Record, 0)
	c := make(chan Record)

	go func() {
		for rec := range c {
			records = append(records, rec)
		}
	}()

	func() {
		defer close(c)

		for _, u := range spider.StartURLs {
			run_scrape(u, c, spider.Parse)
		}
	}()

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(records); err != nil {
		panic(err)
	}
}

func Attr(sel *goquery.Selection, name string) string {
	res, _ := sel.Attr(name)

	return res
}