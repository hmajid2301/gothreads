package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// CollyScraper uses go-colly for robust web scraping
type CollyScraper struct {
	collector *colly.Collector
}

func NewCollyScraper() *CollyScraper {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; GoThreads/1.0)"),
		colly.MaxDepth(1),
		colly.AllowURLRevisit(),
	)

	// Set timeouts
	c.SetRequestTimeout(30 * time.Second)

	// Limit requests
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	return &CollyScraper{collector: c}
}

// ScrapeResult holds the extracted product data
type ScrapeResult struct {
	Title       string
	Description string
	ImageURLs   []string
	SiteName    string
	Price       string
	Brand       string
	Category    string
	Color       string
}

// Scrape extracts product information from a URL using colly
func (s *CollyScraper) Scrape(ctx context.Context, targetURL string) (*ScrapeResult, error) {
	result := &ScrapeResult{
		ImageURLs: []string{},
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	domain := parsedURL.Hostname()

	// Track visited images to avoid duplicates
	seenImages := make(map[string]bool)

	// Extract Open Graph metadata
	s.collector.OnHTML("meta[property]", func(e *colly.HTMLElement) {
		property := e.Attr("property")
		content := e.Attr("content")

		switch property {
		case "og:title":
			if result.Title == "" {
				result.Title = content
			}
		case "og:description":
			if result.Description == "" {
				result.Description = content
			}
		case "og:image":
			if !seenImages[content] {
				seenImages[content] = true
				result.ImageURLs = append(result.ImageURLs, content)
			}
		case "og:site_name":
			if result.SiteName == "" {
				result.SiteName = content
			}
		}
	})

	// Extract Twitter Card metadata (fallback)
	s.collector.OnHTML("meta[name]", func(e *colly.HTMLElement) {
		name := e.Attr("name")
		content := e.Attr("content")

		switch name {
		case "twitter:title":
			if result.Title == "" {
				result.Title = content
			}
		case "twitter:description":
			if result.Description == "" {
				result.Description = content
			}
		case "twitter:image":
			if !seenImages[content] {
				seenImages[content] = true
				result.ImageURLs = append(result.ImageURLs, content)
			}
		}
	})

	// Extract JSON-LD structured data
	s.collector.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(e.Text), &data); err != nil {
			return
		}

		// Extract Product schema
		if dataType, ok := data["@type"].(string); ok {
			if dataType == "Product" || dataType == "IndividualProduct" {
				if result.Title == "" {
					if name, ok := data["name"].(string); ok {
						result.Title = name
					}
				}
				if result.Description == "" {
					if desc, ok := data["description"].(string); ok {
						result.Description = desc
					}
				}
				if result.Brand == "" {
					if brand, ok := data["brand"].(map[string]interface{}); ok {
						if name, ok := brand["name"].(string); ok {
							result.Brand = name
						}
					}
				}
				// Extract images
				switch v := data["image"].(type) {
				case string:
					if !seenImages[v] {
						seenImages[v] = true
						result.ImageURLs = append(result.ImageURLs, v)
					}
				case []interface{}:
					for _, img := range v {
						if imgStr, ok := img.(string); ok && !seenImages[imgStr] {
							seenImages[imgStr] = true
							result.ImageURLs = append(result.ImageURLs, imgStr)
						}
					}
				}
				// Extract price
				if offers, ok := data["offers"].(map[string]interface{}); ok {
					if price, ok := offers["price"].(string); ok {
						currency := ""
						if c, ok := offers["priceCurrency"].(string); ok {
							currency = c
						}
						result.Price = currency + price
					}
				}
			}
		}
	})

	// Fallback to title tag
	s.collector.OnHTML("title", func(e *colly.HTMLElement) {
		if result.Title == "" {
			result.Title = strings.TrimSpace(e.Text)
		}
	})

	// Extract product images from common e-commerce selectors
	s.collector.OnHTML("img", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		dataSrc := e.Attr("data-src")
		dataZoom := e.Attr("data-zoom-image")

		// Check various image attributes
		imgURLs := []string{src, dataSrc, dataZoom}
		for _, imgURL := range imgURLs {
			if imgURL == "" || seenImages[imgURL] {
				continue
			}

			// Skip small images, icons, placeholders
			if strings.Contains(imgURL, "icon") ||
				strings.Contains(imgURL, "logo") ||
				strings.Contains(imgURL, "placeholder") ||
				strings.Contains(imgURL, "spinner") ||
				strings.Contains(imgURL, "loading") ||
				strings.Contains(imgURL, "data:image") {
				continue
			}

			// Make absolute URL
			if !strings.HasPrefix(imgURL, "http") {
				if strings.HasPrefix(imgURL, "//") {
					imgURL = "https:" + imgURL
				} else if strings.HasPrefix(imgURL, "/") {
					imgURL = fmt.Sprintf("https://%s%s", domain, imgURL)
				} else {
					continue
				}
			}

			seenImages[imgURL] = true
			result.ImageURLs = append(result.ImageURLs, imgURL)
		}
	})

	// Site-specific selectors for popular e-commerce sites
	s.setupSiteSpecificHandlers(domain, result)

	// Visit the page
	if err := s.collector.Visit(targetURL); err != nil {
		return nil, fmt.Errorf("failed to scrape: %w", err)
	}

	s.collector.Wait()

	return result, nil
}

// setupSiteSpecificHandlers adds custom handlers for known e-commerce sites
func (s *CollyScraper) setupSiteSpecificHandlers(domain string, result *ScrapeResult) {
	// Uniqlo
	if strings.Contains(domain, "uniqlo") {
		s.collector.OnHTML(".product-information .product-name", func(e *colly.HTMLElement) {
			if result.Title == "" {
				result.Title = strings.TrimSpace(e.Text)
			}
		})
		s.collector.OnHTML(".product-information .product-price", func(e *colly.HTMLElement) {
			if result.Price == "" {
				result.Price = strings.TrimSpace(e.Text)
			}
		})
		s.collector.OnHTML("#fn-product-image img", func(e *colly.HTMLElement) {
			src := e.Attr("src")
			if src != "" && !strings.Contains(src, "placeholder") {
				result.ImageURLs = append(result.ImageURLs, src)
			}
		})
	}

	// Zara
	if strings.Contains(domain, "zara") {
		s.collector.OnHTML("[data-product-name]", func(e *colly.HTMLElement) {
			if result.Title == "" {
				result.Title = e.Attr("data-product-name")
			}
		})
		s.collector.OnHTML(".product-detail-info__name", func(e *colly.HTMLElement) {
			if result.Title == "" {
				result.Title = strings.TrimSpace(e.Text)
			}
		})
	}

	// SSENSE
	if strings.Contains(domain, "ssense") {
		s.collector.OnHTML("h1[data-testid='pdp_product_name']", func(e *colly.HTMLElement) {
			if result.Title == "" {
				result.Title = strings.TrimSpace(e.Text)
			}
		})
		s.collector.OnHTML("[data-testid='pdp_product_brand'] a", func(e *colly.HTMLElement) {
			if result.Brand == "" {
				result.Brand = strings.TrimSpace(e.Text)
			}
		})
	}

	// ASOS
	if strings.Contains(domain, "asos") {
		s.collector.OnHTML("h1[data-auto-id='product-title']", func(e *colly.HTMLElement) {
			if result.Title == "" {
				result.Title = strings.TrimSpace(e.Text)
			}
		})
		s.collector.OnHTML("[data-auto-id='product-price']", func(e *colly.HTMLElement) {
			if result.Price == "" {
				result.Price = strings.TrimSpace(e.Text)
			}
		})
	}

	// Generic price selectors
	s.collector.OnHTML("[itemprop='price']", func(e *colly.HTMLElement) {
		if result.Price == "" {
			result.Price = strings.TrimSpace(e.Text)
		}
	})

	// Generic brand selectors
	s.collector.OnHTML("[itemprop='brand']", func(e *colly.HTMLElement) {
		if result.Brand == "" {
			result.Brand = strings.TrimSpace(e.Text)
		}
	})
}

// ExtractPriceFromText parses price from text using regex
func ExtractPriceFromText(text string) string {
	// Match common price patterns: $50, £30, €25, 50.00 USD, etc.
	re := regexp.MustCompile(`(?:[$£€]|USD|EUR|GBP)?\s*(\d+(?:[.,]\d{2})?)\s*(?:USD|EUR|GBP)?`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// CleanImageURL normalizes image URLs
func CleanImageURL(rawURL, baseURL string) string {
	if strings.HasPrefix(rawURL, "data:") {
		return "" // Skip data URLs
	}

	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}

	if strings.HasPrefix(rawURL, "/") {
		parsed, err := url.Parse(baseURL)
		if err == nil {
			return fmt.Sprintf("https://%s%s", parsed.Hostname(), rawURL)
		}
	}

	return rawURL
}
