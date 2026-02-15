package convert

import (
	"strings"
	"testing"
	"time"
)

func TestStripHidden(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "script tags",
			input: `<div>hello</div><script>alert("xss")</script><p>world</p>`,
			want:  `<div>hello</div><p>world</p>`,
		},
		{
			name:  "script with attributes",
			input: `<div>hello</div><script type="text/javascript">var x=1;</script><p>world</p>`,
			want:  `<div>hello</div><p>world</p>`,
		},
		{
			name:  "style tags",
			input: `<style>.foo { color: red; }</style><p>content</p>`,
			want:  `<p>content</p>`,
		},
		{
			name:  "noscript tags",
			input: `<p>before</p><noscript>Enable JS</noscript><p>after</p>`,
			want:  `<p>before</p><p>after</p>`,
		},
		{
			name:  "template tags",
			input: `<p>before</p><template><div>hidden</div></template><p>after</p>`,
			want:  `<p>before</p><p>after</p>`,
		},
		{
			name:  "HTML comments",
			input: `<p>before</p><!-- this is a comment --><p>after</p>`,
			want:  `<p>before</p><p>after</p>`,
		},
		{
			name:  "multiline comment",
			input: "<p>before</p><!--\nmultiline\ncomment\n--><p>after</p>",
			want:  `<p>before</p><p>after</p>`,
		},
		{
			name:  "hidden attribute",
			input: `<p>visible</p><div hidden>invisible</div><p>also visible</p>`,
			want:  `<p>visible</p><p>also visible</p>`,
		},
		{
			name:  "aria-hidden true",
			input: `<p>visible</p><span aria-hidden="true">invisible</span><p>also visible</p>`,
			want:  `<p>visible</p><p>also visible</p>`,
		},
		{
			name:  "display none",
			input: `<p>visible</p><div style="display:none">invisible</div><p>also visible</p>`,
			want:  `<p>visible</p><p>also visible</p>`,
		},
		{
			name:  "display none with spaces",
			input: `<p>visible</p><div style="display : none">invisible</div><p>also visible</p>`,
			want:  `<p>visible</p><p>also visible</p>`,
		},
		{
			name:  "visibility hidden",
			input: `<p>visible</p><div style="visibility:hidden">invisible</div><p>also visible</p>`,
			want:  `<p>visible</p><p>also visible</p>`,
		},
		{
			name:  "zero-width characters",
			input: "he\u200Bll\u200Co\u200D w\uFEFFor\u2060ld",
			want:  "hello world",
		},
		{
			name:  "bidi overrides",
			input: "hel\u202Alo\u202B wo\u202Crl\u202Dd\u202E!",
			want:  "hello world!",
		},
		{
			name:  "normal HTML unchanged",
			input: `<div><p>Hello <strong>world</strong></p></div>`,
			want:  `<div><p>Hello <strong>world</strong></p></div>`,
		},
		{
			name:  "multiline script",
			input: "<p>before</p><script>\nvar x = 1;\nvar y = 2;\n</script><p>after</p>",
			want:  `<p>before</p><p>after</p>`,
		},
		{
			name:  "onetrust consent banner",
			input: `<p>content</p><div id="onetrust-consent-sdk"><div>We use cookies...</div></div><p>more</p>`,
			want:  `<p>content</p><p>more</p>`,
		},
		{
			name:  "cookiebot banner",
			input: `<p>content</p><div id="CybotCookiebotDialog"><p>cookie prefs</p></div><p>end</p>`,
			want:  `<p>content</p><p>end</p>`,
		},
		{
			name:  "generic cookie-consent id",
			input: `<p>content</p><div id="cookie-consent"><p>consent text</p></div>`,
			want:  `<p>content</p>`,
		},
		{
			name:  "role dialog stripped",
			input: `<p>content</p><div role="dialog" aria-modal="true"><p>Privacy Preferences</p></div><p>end</p>`,
			want:  `<p>content</p><p>end</p>`,
		},
		{
			name:  "role alertdialog stripped",
			input: `<p>content</p><div role="alertdialog"><p>Alert!</p></div><p>end</p>`,
			want:  `<p>content</p><p>end</p>`,
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripHidden(tt.input)
			if got != tt.want {
				t.Errorf("StripHidden()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestStripNav(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "nav element",
			input: `<nav><ul><li>Home</li><li>About</li></ul></nav><main>content</main>`,
			want:  `<main>content</main>`,
		},
		{
			name:  "header element",
			input: `<header><h1>Site Title</h1></header><main>content</main>`,
			want:  `<main>content</main>`,
		},
		{
			name:  "footer element",
			input: `<main>content</main><footer><p>Copyright 2024</p></footer>`,
			want:  `<main>content</main>`,
		},
		{
			name:  "aside element",
			input: `<main>content</main><aside><p>sidebar</p></aside>`,
			want:  `<main>content</main>`,
		},
		{
			name:  "nav with attributes",
			input: `<nav class="main-nav" id="top-nav"><a href="/">Home</a></nav><p>content</p>`,
			want:  `<p>content</p>`,
		},
		{
			name:  "role navigation",
			input: `<div role="navigation"><a href="/">Home</a></div><p>content</p>`,
			want:  `<p>content</p>`,
		},
		{
			name:  "role banner",
			input: `<div role="banner"><h1>Site</h1></div><p>content</p>`,
			want:  `<p>content</p>`,
		},
		{
			name:  "role contentinfo",
			input: `<p>content</p><div role="contentinfo"><p>footer stuff</p></div>`,
			want:  `<p>content</p>`,
		},
		{
			name:  "role complementary",
			input: `<p>content</p><div role="complementary"><p>sidebar</p></div>`,
			want:  `<p>content</p>`,
		},
		{
			name:  "multiple nav elements",
			input: `<nav>top nav</nav><main>content</main><nav>bottom nav</nav>`,
			want:  `<main>content</main>`,
		},
		{
			name:  "normal HTML unchanged",
			input: `<div><p>Hello <strong>world</strong></p></div>`,
			want:  `<div><p>Hello <strong>world</strong></p></div>`,
		},
		{
			name:  "page-level boilerplate stripped around article",
			input: `<header>head</header><main><article><h1>Title</h1><p>Body</p></article></main><footer>foot</footer>`,
			want:  `<main><article><h1>Title</h1><p>Body</p></article></main>`,
		},
		{
			name:  "header inside article preserved",
			input: `<header>site head</header><article><header><h1>Article Title</h1></header><p>Body</p></article>`,
			want:  `<article><header><h1>Article Title</h1></header><p>Body</p></article>`,
		},
		{
			name:  "footer inside article preserved",
			input: `<article><p>Body</p><footer><p>Author bio</p></footer></article><footer>site footer</footer>`,
			want:  `<article><p>Body</p><footer><p>Author bio</p></footer></article>`,
		},
		{
			name:  "aside inside article preserved",
			input: `<aside>sidebar</aside><article><p>Body</p><aside>related content</aside></article>`,
			want:  `<article><p>Body</p><aside>related content</aside></article>`,
		},
		{
			name:  "nav inside article still stripped",
			input: `<article><nav>breadcrumbs</nav><p>Body</p></article>`,
			want:  `<article><p>Body</p></article>`,
		},
		{
			name:  "multiple articles with page boilerplate",
			input: `<header>site</header><article><header>a1</header><p>one</p></article><article><footer>a2</footer><p>two</p></article><footer>site</footer>`,
			want:  `<article><header>a1</header><p>one</p></article><article><footer>a2</footer><p>two</p></article>`,
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripNav(tt.input)
			if got != tt.want {
				t.Errorf("StripNav()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestFormatMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ATX headings preserved",
			input: "# Hello\n\nWorld\n",
			want:  "# Hello\n\nWorld\n",
		},
		{
			name:  "inconsistent spacing normalized",
			input: "#   Hello\n\nSome text\n\n\n\nMore text\n",
			want:  "# Hello\n\nSome text\n\nMore text\n",
		},
		{
			name:  "multiple heading levels",
			input: "# H1\n\n## H2\n\n### H3\n\nText\n",
			want:  "# H1\n\n## H2\n\n### H3\n\nText\n",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "links preserved",
			input: "Visit [example](https://example.com) for more.\n",
			want:  "Visit [example](https://example.com) for more.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatMarkdown(tt.input)
			if err != nil {
				t.Fatalf("FormatMarkdown() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("FormatMarkdown()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestFrontmatter(t *testing.T) {
	m := Metadata{
		SourceURL:   "https://example.com",
		FetchMethod: "browser",
		TimedOut:    false,
		Timing: []TimingStep{
			{"fetch", 1234 * time.Millisecond},
			{"strip_hidden", 1 * time.Millisecond},
			{"convert", 45 * time.Millisecond},
			{"total", 1280 * time.Millisecond},
		},
	}

	got := Frontmatter(m)

	// Check structure.
	if !strings.HasPrefix(got, "---\n") {
		t.Error("frontmatter should start with ---")
	}
	if !strings.HasSuffix(got, "---\n\n") {
		t.Error("frontmatter should end with ---\\n\\n")
	}

	// Check fields present.
	if !strings.Contains(got, "source: https://example.com\n") {
		t.Error("missing source field")
	}
	if !strings.Contains(got, "fetch_method: browser\n") {
		t.Error("missing fetch_method field")
	}
	if !strings.Contains(got, "timed_out: false\n") {
		t.Error("missing timed_out field")
	}

	// Check timing order preserved.
	lines := strings.Split(got, "\n")
	var timingLines []string
	inTiming := false
	for _, line := range lines {
		if line == "timing:" {
			inTiming = true
			continue
		}
		if inTiming && strings.HasPrefix(line, "  ") {
			timingLines = append(timingLines, strings.TrimSpace(line))
		} else if inTiming {
			break
		}
	}

	if len(timingLines) != 4 {
		t.Fatalf("expected 4 timing entries, got %d: %v", len(timingLines), timingLines)
	}
	wantOrder := []string{"fetch:", "strip_hidden:", "convert:", "total:"}
	for i, prefix := range wantOrder {
		if !strings.HasPrefix(timingLines[i], prefix) {
			t.Errorf("timing[%d]: want prefix %q, got %q", i, prefix, timingLines[i])
		}
	}
}

func TestFrontmatterTimedOut(t *testing.T) {
	m := Metadata{
		SourceURL:   "https://slow.example.com",
		FetchMethod: "browser",
		TimedOut:    true,
	}
	got := Frontmatter(m)
	if !strings.Contains(got, "timed_out: true\n") {
		t.Error("timed_out should be true")
	}
	// No timing section when empty.
	if strings.Contains(got, "timing:") {
		t.Error("should not have timing section when no timing steps")
	}
}
