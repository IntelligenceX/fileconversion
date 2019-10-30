/*
File Name:  HTML 2 Text.go
Copyright:  2018 Kleissner Investments s.r.o.
Author:     Peter Kleissner
*/

package fileconversion

import (
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/IntelligenceX/fileconversion/html2text"
	"github.com/PuerkitoBio/goquery"
	"github.com/ssor/bom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// HTML2Text extracts the text from the HTML
func HTML2Text(reader io.Reader) (pageText string, err error) {
	// The charset.NewReader ensures that foreign encodings are properly decoded to UTF-8.
	// It will make both heuristic checks as well as look for the HTML meta charset tag.
	reader, err = charset.NewReader(reader, "")
	if err != nil {
		return "", err
	}

	// The html2text is a forked improved version that converts HTML to human-friendly text.
	return html2text.FromReader(reader)
}

// HTML2TextAndLinks extracts the text from the HTML and all links from <a> and <img> tags of a HTML
// If the base URL is provided, relative links will be converted to absolute ones.
func HTML2TextAndLinks(reader io.Reader, baseURL string) (pageText string, links []string, err error) {
	// The charset.NewReader ensures that foreign encodings are properly decoded to UTF-8.
	// It will make both heuristic checks as well as look for the HTML meta charset tag.
	reader, err = charset.NewReader(reader, "")
	if err != nil {
		return "", nil, err
	}

	// code from html2text.FromReader to parse the doc
	newReader, err := bom.NewReaderWithoutBom(reader)
	if err != nil {
		return "", nil, err
	}
	doc, err := html.Parse(newReader)
	if err != nil {
		return "", nil, err
	}

	// get the text
	pageText, err = html2text.FromHTMLNode(doc)
	if err != nil {
		return pageText, nil, err
	}

	// get the links
	docQ := goquery.NewDocumentFromNode(doc)
	docQ.Url, _ = url.Parse(baseURL)
	links = processLinks(docQ)

	return pageText, links, err
}

// ---- below 2 functions are forks from gocrawl/worker.go ----

func handleBaseTag(root *url.URL, baseHref string, aHref string) string {
	resolvedBase, err := root.Parse(baseHref)
	if err != nil {
		return ""
	}

	parsedURL, err := url.Parse(aHref)
	if err != nil {
		return ""
	}
	// If a[href] starts with a /, it overrides the base[href]
	if parsedURL.Host == "" && !strings.HasPrefix(aHref, "/") {
		aHref = path.Join(resolvedBase.Path, aHref)
	}

	resolvedURL, err := resolvedBase.Parse(aHref)
	if err != nil {
		return ""
	}
	return resolvedURL.String()
}

// Scrape the document's content to gather all links
func processLinks(doc *goquery.Document) (result []string) {
	// process links via <a href=""> tags
	baseURL, _ := doc.Find("base[href]").Attr("href")
	urls := doc.Find("a[href]").Map(func(_ int, s *goquery.Selection) string {
		val, _ := s.Attr("href")
		if baseURL != "" {
			val = handleBaseTag(doc.Url, baseURL, val)
		}
		return val
	})

	// all image references via <img src=""> tag
	imgURLs := doc.Find("img[src]").Map(func(_ int, s *goquery.Selection) string {
		val, _ := s.Attr("src")
		if baseURL != "" {
			val = handleBaseTag(doc.Url, baseURL, val)
		}
		return val
	})
	urls = append(urls, imgURLs...)

	// form submission links <form action="/action_page.php" method="get">
	formURLs := doc.Find("form[action]").Map(func(_ int, s *goquery.Selection) string {
		val, _ := s.Attr("action")
		if baseURL != "" {
			val = handleBaseTag(doc.Url, baseURL, val)
		}
		return val
	})
	urls = append(urls, formURLs...)

	// parse all found URLs
	for _, s := range urls {
		// If href starts with "#", then it points to this same exact URL, ignore (will fail to parse anyway)
		if len(s) > 0 && !strings.HasPrefix(s, "#") {
			if parsed, e := url.Parse(s); e == nil {
				parsed = doc.Url.ResolveReference(parsed)

				result = append(result, parsed.String())
				//fmt.Printf("%s\n", parsed.String())
			} else {
				//w.logFunc(LogIgnored, "ignore on unparsable policy %s: %s", s, e.Error())
			}
		}
	}
	return
}
