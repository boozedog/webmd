package preview

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
)

// Render converts markdown to a complete HTML page.
func Render(markdown string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}

	var out bytes.Buffer
	out.WriteString(`<!DOCTYPE html>
<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>webmd preview</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Atkinson+Hyperlegible+Next:ital,wght@0,400;0,700;1,400;1,700&display=swap" rel="stylesheet">
<style>
body { max-width: 800px; margin: 40px auto; padding: 0 20px;
  font-family: "Atkinson Hyperlegible Next", -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
  line-height: 1.6; color: #333; background: #fff; }
h1, h2, h3 { margin-top: 1.5em; }
code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; font-size: 0.9em; }
pre { background: #f4f4f4; padding: 16px; border-radius: 6px; overflow-x: auto; }
pre code { background: none; padding: 0; }
blockquote { border-left: 4px solid #ddd; margin-left: 0; padding-left: 16px; color: #666; }
img { max-width: 100%; }
a { color: #0366d6; }
table { border-collapse: collapse; }
th, td { border: 1px solid #ddd; padding: 8px 12px; }
@media (prefers-color-scheme: dark) {
  body { background: #1a1a1a; color: #d4d4d4; }
  code, pre { background: #2d2d2d; }
  blockquote { border-color: #555; color: #aaa; }
  a { color: #58a6ff; }
  th, td { border-color: #444; }
}
</style>
</head><body>
`)
	buf.WriteTo(&out)
	out.WriteString("\n</body></html>\n")

	return out.String(), nil
}
