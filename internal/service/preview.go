package service

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	reFencedCode  = regexp.MustCompile("(?s)```[^\\n]*\\n?(.*?)```")
	reInlineCode  = regexp.MustCompile("`([^`]+)`")
	reImage       = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	reLink        = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	reRefLinkDef  = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s+\S+.*$`)
	reRefLink     = regexp.MustCompile(`\[([^\]]+)\]\[[^\]]*\]`)
	reHeading     = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	reBlockquote  = regexp.MustCompile(`(?m)^>\s*`)
	reListMarker  = regexp.MustCompile(`(?m)^\s*(?:[-*+]|\d+\.)\s+`)
	reHR          = regexp.MustCompile(`(?m)^\s*(?:[-*_]){3,}\s*$`)
	reHTMLTag     = regexp.MustCompile(`<[^>]+>`)
	reBoldItalic  = regexp.MustCompile(`(?s)(\*\*\*|___)(.+?)(\*\*\*|___)`)
	reBold        = regexp.MustCompile(`(?s)(\*\*|__)(.+?)(\*\*|__)`)
	reStrike      = regexp.MustCompile(`(?s)~~(.+?)~~`)
	reItalicStar  = regexp.MustCompile(`\*([^*\n]+)\*`)
	reItalicUnder = regexp.MustCompile(`(^|[^\pL\pN])_([^_\n]+)_([^\pL\pN]|$)`)
	reEscaped     = regexp.MustCompile(`\\([\\` + "`" + `*_\{}\[\]()#+\-.!>~|])`)
	reMultiSpace  = regexp.MustCompile(`\s+`)
)

func StripMarkdown(md string) string {
	s := md
	s = reFencedCode.ReplaceAllString(s, " $1 ")
	s = reInlineCode.ReplaceAllString(s, "$1")
	s = reImage.ReplaceAllString(s, "$1")
	s = reLink.ReplaceAllString(s, "$1")
	s = reRefLinkDef.ReplaceAllString(s, "")
	s = reRefLink.ReplaceAllString(s, "$1")
	s = reHR.ReplaceAllString(s, " ")
	s = reHeading.ReplaceAllString(s, "")
	s = reBlockquote.ReplaceAllString(s, "")
	s = reListMarker.ReplaceAllString(s, "")
	s = reHTMLTag.ReplaceAllString(s, " ")
	s = reBoldItalic.ReplaceAllString(s, "$2")
	s = reBold.ReplaceAllString(s, "$2")
	s = reStrike.ReplaceAllString(s, "$1")
	s = reItalicStar.ReplaceAllString(s, "$1")
	s = reItalicUnder.ReplaceAllString(s, "$1$2$3")
	s = reEscaped.ReplaceAllString(s, "$1")
	s = reMultiSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func GeneratePreview(content string, maxRunes int) string {
	plain := StripMarkdown(content)
	if maxRunes <= 0 {
		maxRunes = 250
	}
	if utf8.RuneCountInString(plain) <= maxRunes {
		return plain
	}

	runes := []rune(plain)
	truncated := string(runes[:maxRunes])
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxRunes/2 {
		truncated = truncated[:lastSpace]
	}
	return strings.TrimSpace(truncated) + "..."
}
