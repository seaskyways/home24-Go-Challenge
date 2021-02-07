package webalayzer

import (
	"context"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

type Analyzer struct {
	// checking inaccessible links is very taxing.
	// This option allows the user to skip checking in/accessible links
	SkipInaccessibleCheck bool

	// empty Anchors are often used to return the user to the homepage.
	// this option allows the user to choose whether these should be counted or not
	SkipEmptyLinks bool
}

func (a *Analyzer) AnalyzeURL(ctx context.Context, url string) (stats WebPageStats, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.Body == nil {
		err = ErrBadURL
		return
	}
	defer resp.Body.Close()

	ctx = context.WithValue(context.Background(), "url", url)

	return a.AnalyzeReader(ctx, resp.Body)
}

func (a *Analyzer) AnalyzeReader(ctx context.Context, reader io.Reader) (stats WebPageStats, err error) {
	startTime := time.Now()
	defer func() { stats.ProcessingDuration = time.Now().Sub(startTime) }()

	node, err := html.Parse(reader)
	if err != nil {
		return
	}

	// defaults to 5.0
	stats.HTMLVersion = "5.0"

	if u, ok := ctx.Value("url").(string); ok {
		parsedUrl, _ := url.Parse(u)
		ctx = context.WithValue(ctx, "parsedUrl", parsedUrl)
	}

	wpsCtx := &WebPageStatsContext{
		Context:      ctx,
		WebPageStats: &stats,
	}
	err = traverseNode(wpsCtx, node, a.analyzeNode)

	// wait for all accessibility checkers
	wpsCtx.accessibilityWg.Wait()
	return
}

func (a *Analyzer) analyzeNode(ctx context.Context, node *html.Node) error {
	if node == nil || ctx.Err() != nil {
		return nil
	}

	statsCtx := ctx.(*WebPageStatsContext)

	switch node.Type {
	case html.DoctypeNode:
		statsCtx.HTMLVersion = a.getHTMLVersion(node.Attr)

	case html.ElementNode:
		if err := a.analyzeElementNode(node, statsCtx); err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) analyzeElementNode(node *html.Node, stats *WebPageStatsContext) error {
	switch node.Data {
	case "title":
		if node.FirstChild != nil && node.FirstChild.Type == html.TextNode {
			stats.Title = node.FirstChild.Data
		}

	case "h1":
		stats.H1Count++
	case "h2":
		stats.H2Count++
	case "h3":
		stats.H3Count++
	case "h4":
		stats.H4Count++
	case "h5":
		stats.H5Count++
	case "h6":
		stats.H6Count++

	case "a":
		if attr := findAttributeByKey(node.Attr, "href"); attr != nil {
			if err := a.analyzeLink(attr.Val, stats); err != nil {
				fmt.Printf("error analyzing anchor link (%v): %v", attr.Val, err)
			}
			if !a.SkipInaccessibleCheck {
				if err := a.analyzeLinkAccessible(attr.Val, stats); err != nil {
					fmt.Printf("error analyzing link accessibility (%v): %v", attr.Val, err)
				}
			}
		}

	case "form":
		if attr := findAttributeByKey(node.Attr, "action"); attr != nil {
			if err := a.analyzeLink(attr.Val, stats); err != nil {
				fmt.Printf("error analyzing form link (%v): %v", attr.Val, err)
			}
		}

	case "input":
		if checkAttributeEquals(node, "type", "password") {
			stats.PasswordInputsCount++
		}
	}
	return nil
}

func (a *Analyzer) analyzeLink(link string, stats *WebPageStatsContext) error {
	if len(link) == 0 {
		if !a.SkipEmptyLinks {
			stats.InternalLinksCount++
		}
		return nil
	}

	var pageHost string
	if parsedUrl, ok := stats.Value("parsedUrl").(*url.URL); ok {
		pageHost = parsedUrl.Host
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return err
	}

	if !parsed.IsAbs() || parsed.Host == pageHost {
		stats.InternalLinksCount++
	} else {
		stats.ExternalLinksCount++
	}

	return nil
}

func (a *Analyzer) analyzeLinkAccessible(link string, stats *WebPageStatsContext) error {
	if len(link) == 0 {
		if !a.SkipEmptyLinks {
			stats.AccessibleLinksCount++
		}
		return nil
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return err
	}

	// branch accessibility checkers to their own goroutines to speed things up
	stats.accessibilityWg.Add(1)
	go func(link string) {
		defer stats.accessibilityWg.Done()

		if !parsed.IsAbs() {
			// add relative URL to current page to make the link absolute
			var currentPage string
			if parsedUrl, ok := stats.Value("parsedUrl").(*url.URL); ok && parsedUrl != nil {
				currentPage = parsedUrl.String()
			}
			link = currentPage + "/" + link
		}

		// using HEAD instead of GET to avoid downloading data we don't need
		// From Mozilla Docs: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD
		// The HTTP HEAD method requests the headers that would be returned
		// if the HEAD request's URL was instead requested with the HTTP GET method.
		resp, err := http.Head(link)
		if err != nil || resp == nil || resp.StatusCode >= 400 || resp.StatusCode < 200 {
			atomic.AddInt32(&stats.InaccessibleLinksCount, 1)
		} else {
			atomic.AddInt32(&stats.AccessibleLinksCount, 1)
		}
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}(link)

	return nil
}

func (_ *Analyzer) getHTMLVersion(doctypeAttrs []html.Attribute) string {
	if attr := findAttributeByKey(doctypeAttrs, "public"); attr != nil && strings.Contains(attr.Val, "4.0") {
		return "4.0"
	} else {
		return "5.0"
	}
}

func findAttributeByKey(attrs []html.Attribute, key string) *html.Attribute {
	for i := range attrs {
		if attrs[i].Key == key {
			return &attrs[i]
		}
	}
	return nil
}

func checkAttributeEquals(node *html.Node, key, value string) bool {
	if attr := findAttributeByKey(node.Attr, key); attr != nil && attr.Val == value {
		return true
	}
	return false
}
