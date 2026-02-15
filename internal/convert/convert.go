package convert

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/mackee/go-readability"
	goldmarkmd "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
)

var imgTagRe = regexp.MustCompile(`<img\b[^>]*/?>`)
var junkLinkRe = regexp.MustCompile(`\[([^\]]*)\]\((#[^)]*|\s*)?\)`)

// StripHidden regexes — non-visible content that adds noise and prompt injection risk.
var (
	scriptRe    = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	styleRe     = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	noscriptRe  = regexp.MustCompile(`(?is)<noscript\b[^>]*>.*?</noscript>`)
	templateRe  = regexp.MustCompile(`(?is)<template\b[^>]*>.*?</template>`)
	commentRe   = regexp.MustCompile(`(?s)<!--.*?-->`)
	zeroWidthRe = regexp.MustCompile("[\u200B\u200C\u200D\uFEFF\u2060\u202A\u202B\u202C\u202D\u202E]")

	// Opening-tag matchers for attribute-based stripping (capture group 1 = tag name).
	hiddenOpenRe           = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\bhidden\b[^>]*>`)
	ariaHiddenOpenRe       = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\baria-hidden\s*=\s*"true"[^>]*>`)
	displayNoneOpenRe      = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\bstyle\s*=\s*"[^"]*display\s*:\s*none[^"]*"[^>]*>`)
	visibilityHiddenOpenRe = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\bstyle\s*=\s*"[^"]*visibility\s*:\s*hidden[^"]*"[^>]*>`)

	// Cookie/consent banner matchers — common consent SDK wrapper IDs and role="dialog" modals.
	consentIDOpenRe  = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\bid\s*=\s*"(onetrust-consent-sdk|cookiebot|CybotCookiebotDialog|cookie-consent|cookie-banner|cookie-notice|consent-banner|gdpr-consent|cc-window|cc_div)"[^>]*>`)
	roleDialogOpenRe = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\brole\s*=\s*"(dialog|alertdialog)"[^>]*>`)
)

// StripNav regexes — semantic boilerplate elements.
var (
	// Nav is always stripped everywhere.
	navRe         = regexp.MustCompile(`(?is)<nav\b[^>]*>.*?</nav>`)
	roleNavOpenRe = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\brole\s*=\s*"navigation"[^>]*>`)

	// These are only stripped outside <article> — inside articles they're real content.
	headerRe = regexp.MustCompile(`(?is)<header\b[^>]*>.*?</header>`)
	footerRe = regexp.MustCompile(`(?is)<footer\b[^>]*>.*?</footer>`)
	asideRe  = regexp.MustCompile(`(?is)<aside\b[^>]*>.*?</aside>`)

	roleBannerOpenRe        = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\brole\s*=\s*"banner"[^>]*>`)
	roleContentInfoOpenRe   = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\brole\s*=\s*"contentinfo"[^>]*>`)
	roleComplementaryOpenRe = regexp.MustCompile(`(?i)<(\w+)\b[^>]*\brole\s*=\s*"complementary"[^>]*>`)

	articleOpenRe = regexp.MustCompile(`(?i)<article\b[^>]*>`)
	articleRe     = regexp.MustCompile(`(?is)<article\b[^>]*>.*?</article>`)
)

// stripByAttr finds elements whose opening tag matches openRe (which must capture the tag name
// in group 1), then removes everything from the opening tag through the balanced closing tag.
func stripByAttr(html string, openRe *regexp.Regexp) string {
	for {
		m := openRe.FindStringSubmatchIndex(html)
		if m == nil {
			break
		}
		tagName := strings.ToLower(html[m[2]:m[3]])
		closeTag := "</" + tagName + ">"
		openTag := "<" + tagName

		rest := html[m[1]:]
		restLower := strings.ToLower(rest)
		depth := 1
		pos := 0
		for depth > 0 {
			closeIdx := strings.Index(restLower[pos:], closeTag)
			if closeIdx == -1 {
				// No balanced close — strip just the opening tag.
				pos = -1
				break
			}
			// Count any nested opens of the same tag between pos and closeIdx.
			segment := restLower[pos : pos+closeIdx]
			for search := 0; ; {
				idx := strings.Index(segment[search:], openTag)
				if idx == -1 {
					break
				}
				// Verify it's a real tag open (followed by space, >, or /).
				next := search + idx + len(openTag)
				if next < len(segment) {
					ch := segment[next]
					if ch == ' ' || ch == '>' || ch == '/' || ch == '\t' || ch == '\n' {
						depth++
					}
				}
				search = search + idx + len(openTag)
			}
			depth--
			pos = pos + closeIdx + len(closeTag)
		}

		if pos == -1 {
			html = html[:m[0]] + html[m[1]:]
		} else {
			html = html[:m[0]] + html[m[1]+pos:]
		}
	}
	return html
}

// StripHidden removes non-visible HTML content: script/style/noscript/template tags,
// HTML comments, hidden/aria-hidden elements, display:none/visibility:hidden elements,
// cookie/consent banners, modal dialogs, and zero-width Unicode characters.
func StripHidden(html string) string {
	html = scriptRe.ReplaceAllString(html, "")
	html = styleRe.ReplaceAllString(html, "")
	html = noscriptRe.ReplaceAllString(html, "")
	html = templateRe.ReplaceAllString(html, "")
	html = commentRe.ReplaceAllString(html, "")
	html = stripByAttr(html, hiddenOpenRe)
	html = stripByAttr(html, ariaHiddenOpenRe)
	html = stripByAttr(html, displayNoneOpenRe)
	html = stripByAttr(html, visibilityHiddenOpenRe)
	html = stripByAttr(html, consentIDOpenRe)
	html = stripByAttr(html, roleDialogOpenRe)
	html = zeroWidthRe.ReplaceAllString(html, "")
	return html
}

// StripNav removes semantic navigation/boilerplate elements.
// Nav (and role="navigation") is always stripped everywhere.
// Header, footer, aside (and role banner/contentinfo/complementary) are only stripped
// outside <article> elements — inside articles they represent real content.
func StripNav(html string) string {
	// Always strip nav everywhere.
	html = navRe.ReplaceAllString(html, "")
	html = stripByAttr(html, roleNavOpenRe)

	// For header/footer/aside: protect <article> content by replacing articles with
	// placeholders, stripping boilerplate from the rest, then restoring articles.
	if !articleOpenRe.MatchString(html) {
		// No articles — strip everything directly.
		html = stripBoilerplate(html)
		return html
	}

	articles := articleRe.FindAllString(html, -1)
	const placeholder = "\x00ARTICLE\x00"
	shell := articleRe.ReplaceAllString(html, placeholder)
	shell = stripBoilerplate(shell)
	for _, a := range articles {
		shell = strings.Replace(shell, placeholder, a, 1)
	}
	return shell
}

// stripBoilerplate removes header, footer, aside and their ARIA role equivalents.
func stripBoilerplate(html string) string {
	html = headerRe.ReplaceAllString(html, "")
	html = footerRe.ReplaceAllString(html, "")
	html = asideRe.ReplaceAllString(html, "")
	html = stripByAttr(html, roleBannerOpenRe)
	html = stripByAttr(html, roleContentInfoOpenRe)
	html = stripByAttr(html, roleComplementaryOpenRe)
	return html
}

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

// FormatMarkdown normalizes markdown by parsing it through goldmark and re-rendering
// with the goldmark-markdown renderer (ATX headings, consistent whitespace).
func FormatMarkdown(md string) (string, error) {
	renderer := goldmarkmd.NewRenderer()
	gm := goldmark.New(goldmark.WithRenderer(renderer))

	var buf bytes.Buffer
	if err := gm.Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("formatting markdown: %w", err)
	}
	return buf.String(), nil
}

// TimingStep records the duration of a single pipeline step.
type TimingStep struct {
	Name     string
	Duration time.Duration
}

// Metadata holds information about a fetch for frontmatter generation.
type Metadata struct {
	SourceURL   string
	FetchMethod string // "markdown" or "browser"
	TimedOut    bool
	Timing      []TimingStep
}

// Frontmatter generates a YAML frontmatter block from metadata.
func Frontmatter(m Metadata) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "source: %s\n", m.SourceURL)
	fmt.Fprintf(&b, "fetch_method: %s\n", m.FetchMethod)
	fmt.Fprintf(&b, "timed_out: %t\n", m.TimedOut)
	if len(m.Timing) > 0 {
		b.WriteString("timing:\n")
		for _, step := range m.Timing {
			fmt.Fprintf(&b, "  %s: %s\n", step.Name, step.Duration.Round(time.Millisecond))
		}
	}
	b.WriteString("---\n\n")
	return b.String()
}
