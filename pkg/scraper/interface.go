package scraper

type BrowserScraper interface {
	Scrape(url string) ([]string, error)
}
