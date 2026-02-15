package convert

import (
	"fmt"
	"regexp"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/mackee/go-readability"
)

var imgTagRe = regexp.MustCompile(`<img\b[^>]*/?>`)
var junkLinkRe = regexp.MustCompile(`\[([^\]]*)\]\((#[^)]*|\s*)?\)`)

// StripImages removes <img> tags from HTML.
func StripImages(html string) string {
	return imgTagRe.ReplaceAllString(html, "")
}

// StripJunkLinks removes empty links [text]() and anchor-only links [text](#foo) from markdown,
// replacing them with just their text content.
func StripJunkLinks(md string) string {
	return junkLinkRe.ReplaceAllString(md, "")
}

// Readability extracts the main content from HTML and converts it to markdown.
// Falls back to Full() if readability cannot extract an article.
func Readability(html string) (string, error) {
	article, err := readability.Extract(html, readability.DefaultOptions())
	if err != nil {
		return Full(html)
	}

	if article.Root == nil {
		return Full(html)
	}

	body := strings.TrimSpace(readability.ToMarkdown(article.Root))
	if body == "" {
		return Full(html)
	}

	var b strings.Builder
	if article.Title != "" {
		fmt.Fprintf(&b, "# %s\n\n", article.Title)
	}
	if article.Byline != "" {
		fmt.Fprintf(&b, "*%s*\n\n", article.Byline)
	}

	b.WriteString(body)
	b.WriteByte('\n')

	return b.String(), nil
}

// Full converts the entire HTML page to markdown.
func Full(html string) (string, error) {
	md, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("converting HTML to markdown: %w", err)
	}
	return md, nil
}
