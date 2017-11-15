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

func get_doc_from_filename(fname string) (*goquery.Document, error) {
	f, err := os.OpenFile(fname, os.O_RDONLY, 0444)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return goquery.NewDocumentFromReader(f)
}

func get_doc_from_url(urlString string) (*goquery.Document, error) {
	return goquery.NewDocument(urlString)
}

func run_scrape(urlString string, c chan<- Record, parser ParseFunc) error {
	u, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	var doc_getter func(string) (*goquery.Document, error)

	if u.Scheme == "file" {
		doc_getter = get_doc_from_filename
	} else {
		doc_getter = get_doc_from_url
	}

	doc, err := doc_getter(urlString)
	if err != nil {
		return err
	}

	parser(doc, c)

	return nil
}

func RunSpider(spider Spider) {
	records := make([]Record, 0)
	c := make(chan Record)

	go func() {
		for rec := range c {
			records = append(records, rec)
		}
	}()

	for _, u := range spider.StartURLs {
		if err := run_scrape(u, c, spider.Parse); err != nil {
			panic(err)
		}
	}

	close(c)

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(records); err != nil {
		panic(err)
	}
}

func Attr(sel *goquery.Selection, name string) string {
	res, _ := sel.Attr(name)

	return res
}