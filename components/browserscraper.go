package components

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"insta-scraper/pkg/scraper"

	"github.com/chromedp/chromedp"
)

var _ scraper.BrowserScraper = (*BrowserScraper)(nil)

type BrowserScraper struct {
	browserType string
	browserPath string
	timeout     time.Duration
}

func NewBrowserScraper(browserType, browserPath string) *BrowserScraper {
	return &BrowserScraper{
		browserType: browserType,
		browserPath: browserPath,
		timeout:     30 * time.Second,
	}
}

func (bs *BrowserScraper) Scrape(url string) ([]string, error) {
	log.Printf("Starting browser scrape for: %s", url)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(bs.browserPath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := context.WithTimeout(allocCtx, bs.timeout)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// JavaScript to scrape Instagram media
	script := `
		(function() {
			const results = [];
			const articles = Array.from(document.querySelectorAll('article'));
			
			for (const article of articles) {
				const postLink = article.querySelector('a[href*="/p/"]');
				if (!postLink) continue;
				
				const mediaElements = Array.from(article.querySelectorAll('img, video'));
				for (const el of mediaElements) {
					if (el.src) {
						results.push(el.src);
					}
				}
			}
			return results;
		})()
	`

	var urls []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second), // Reduced wait time
		chromedp.WaitReady(`article`), // Wait specifically for posts to load
		chromedp.Evaluate(script, &urls),
	)
	if err != nil {
		return nil, fmt.Errorf("browser scrape failed: %w", err)
	}

	// Filter and validate URLs
	var filtered []string
	for _, url := range urls {
		if strings.HasPrefix(url, "http") {
			filtered = append(filtered, url)
			log.Printf("Found media URL: %s", url)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no valid media URLs found")
	}

	log.Printf("Scraped %d media URLs successfully", len(filtered))
	return filtered, nil
}
