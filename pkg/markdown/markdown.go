package markdown

import (
	"html"
	"regexp"
	"strings"
)

var (
	codePattern = regexp.MustCompile("`([^`]+)`")
	boldPattern = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	linkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
)

// Render turns a small slice of Markdown into HTML. The source is escaped before any tags are added.
func Render(src string) string {
	var b strings.Builder
	inCode := false
	inList := false
	closeList := func() {
		if inList {
			b.WriteString("</ul>")
			inList = false
		}
	}
	for _, line := range strings.Split(src, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			closeList()
			if inCode {
				b.WriteString("</code></pre>")
				inCode = false
			} else {
				b.WriteString(`<pre class="overflow-auto rounded bg-gray-50 p-3 text-sm"><code>`)
				inCode = true
			}
			continue
		}
		if inCode {
			b.WriteString(html.EscapeString(line))
			b.WriteByte('\n')
			continue
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			if !inList {
				b.WriteString(`<ul class="list-disc space-y-1 pl-5">`)
				inList = true
			}
			b.WriteString("<li>")
			b.WriteString(inline(line[2:]))
			b.WriteString("</li>")
			continue
		}
		closeList()
		switch {
		case strings.HasPrefix(line, "### "):
			b.WriteString(`<h3 class="text-base font-semibold text-gray-900">`)
			b.WriteString(inline(line[4:]))
			b.WriteString("</h3>")
		case strings.HasPrefix(line, "## "):
			b.WriteString(`<h2 class="text-lg font-semibold text-gray-900">`)
			b.WriteString(inline(line[3:]))
			b.WriteString("</h2>")
		case strings.HasPrefix(line, "# "):
			b.WriteString(`<h1 class="text-xl font-semibold text-gray-900">`)
			b.WriteString(inline(line[2:]))
			b.WriteString("</h1>")
		case trim == "":
		default:
			b.WriteString(`<p class="text-sm text-gray-800">`)
			b.WriteString(inline(line))
			b.WriteString("</p>")
		}
	}
	closeList()
	if inCode {
		b.WriteString("</code></pre>")
	}
	return b.String()
}

func inline(src string) string {
	escaped := html.EscapeString(src)
	escaped = codePattern.ReplaceAllString(escaped, `<code class="rounded bg-gray-100 px-1">$1</code>`)
	escaped = boldPattern.ReplaceAllString(escaped, "<strong>$1</strong>")
	return linkPattern.ReplaceAllStringFunc(escaped, func(match string) string {
		parts := linkPattern.FindStringSubmatch(match)
		if len(parts) != 3 || !safeURL(parts[2]) {
			return parts[0]
		}
		return `<a class="text-violet-800 underline" href="` + parts[2] + `">` + parts[1] + `</a>`
	})
}

func safeURL(raw string) bool {
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") {
		return false
	}
	return strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") || strings.HasPrefix(raw, "/")
}
