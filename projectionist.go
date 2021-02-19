package proj

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var ProjectionFile = ".projections.json"

type Props map[string]string
type Projection map[string]Props

func Project(filename string) Props {
	return Props(map[string]string{})
}

var matchSplit *regexp.Regexp
var loneStar *regexp.Regexp

func init() {
	matchSplit = regexp.MustCompile(`\*\*?`)
	loneStar = regexp.MustCompile(`^[^*{}]*\*[^*{}]*$`)
}

func matches(pattern, filename string) (string, bool) {
	if loneStar.MatchString(pattern) {
		pattern = strings.Replace(pattern, "*", "**/*", 1)
	}
	comp := matchSplit.Split(pattern, -1)
	if len(comp) != 3 {
		panic("Should have splet the path in 3")
	}
	if !strings.HasPrefix(filename, comp[0]) || !strings.HasSuffix(filename, comp[2]) {
		return "", false
	}
	match := filename[len(comp[0]) : len(filename)-len(comp[2])]
	if comp[1] == "/" {
		return path.Clean(match), true
	}
	clean := regexp.MustCompile(
		`(?i)`+regexp.QuoteMeta(comp[1])+`([^/]*$)`,
	).ReplaceAllString("/"+match, "/$1")[1:]
	return path.Clean(clean), true
}

func Detect(filename string) ([]string, error) {
	var projections = []string{}
	wd, err := os.Getwd()
	if err != nil {
		return projections, err
	}
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(wd, filename)
	}
	dir := filename
	stat, err := os.Stat(dir)
	if err != nil {
		return projections, err
	}
	if !stat.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		name := filepath.Join(dir, ProjectionFile)
		f, err := os.Open(name)
		if err != nil {
			goto Next
		}
		if err := f.Close(); err != nil {
			goto Next
		}
		projections = append(projections, name)
	Next:
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return projections, nil
}

func Dot(s string) string {
	return strings.Replace(s, "/", ".", -1)
}

func Underscore(s string) string {
	return strings.Replace(s, "/", "_", -1)
}

func Backslash(s string) string {
	return strings.Replace(s, "/", `\`, -1)
}

func Colons(s string) string {
	return strings.Replace(s, "/", "::", -1)
}

func Hyphenate(s string) string {
	return strings.Replace(s, "_", "-", -1)
}

var blankReplacer *strings.Replacer

func init() {
	blankReplacer = strings.NewReplacer("_", " ", "-", " ")
}

func Blank(s string) string {
	return blankReplacer.Replace(s)
}

func UpperCase(s string) string {
	return strings.ToUpper(s)
}

func CamelCase(s string) string {
	var b strings.Builder
	n := strings.Count(s, "_")
	b.Grow(len(s) - n)
	start := 0
	for i := 0; i < n; i++ {
		j := start
		j += strings.Index(s[start:], "_")
		b.WriteString(s[start:j])
		if j+1 < len(s) {
			b.WriteString(strings.ToUpper(s[j+1 : j+2]))
			start = j + 2
		} else {
			start = j + 1
		}
	}
	b.WriteString(s[start:])
	return b.String()
}

func SnakeCase(s string) string {
	var b strings.Builder
	b.Grow(len(s) * 2)

	start := 0
	for i, r := range s {
		if unicode.IsUpper(r) {
			b.WriteString(s[start:i])
			b.WriteRune('_')
			b.WriteRune(unicode.ToLower(r))
			start = i + 1
		}
	}
	b.WriteString(s[start:])

	return b.String()
}

func Capitalize(s string) string {
	var b strings.Builder

	if len(s) == 0 {
		return s
	}

	b.Grow(len(s))
	rs := []rune(s)
	b.WriteRune(unicode.ToUpper(rs[0]))
	start := 1
	for i := 1; i < len(rs); i++ {
		if rs[i] == '/' {
			b.WriteString(string(rs[start : i+1]))
			i++
			if i+1 < len(rs) {
				b.WriteRune(unicode.ToUpper(rs[i+1]))
				i++
			}
			start = i
		}
	}
	b.WriteString(string(rs[start:]))
	return b.String()
}

func Dirname(s string) string {
	return path.Dir(s)
}

func Basename(s string) string {
	return path.Base(s)
}

func replaceLookingBack(
	s, lookback, pattern, replacement string,
	match bool,
) string {
	p := regexp.MustCompile(pattern)
	l := regexp.MustCompile(lookback)
	loc := p.FindStringIndex(s)
	if loc == nil {
		return s
	}
	m := l.MatchString(s[:loc[0]])
	if m == match {
		var b strings.Builder
		b.Grow(len(s[:loc[0]]) + len(replacement))
		b.WriteString(s[:loc[0]])
		b.WriteString(replacement)
		b.WriteString(s[loc[1]:])
		return b.String()
	}
	return s
}

func Singular(s string) string {
	s = replaceLookingBack(
		s, "[Mm]ov|[aeio]$", "ies$", "ys",
		false,
	)
	s = replaceLookingBack(
		s, "[rl]$", "ves$", "fs",
		true,
	)
	s = replaceLookingBack(
		s, "(nd|rt)$", "ices$", "exs",
		true,
	)
	s = replaceLookingBack(
		s, "s$", "s$", "",
		false,
	)
	s = replaceLookingBack(
		s, "[nrt]ch|tatus|lias|ss$", "e$", "",
		true,
	)
	return s
}

func Plural(s string) string {
	s = replaceLookingBack(
		s, "[aeio]$", "y$", "ie",
		false,
	)
	s = replaceLookingBack(
		s, "[rl]$", "f$", "ve",
		true,
	)
	s = replaceLookingBack(
		s, "nd|rt$", "ex$", "ice",
		true,
	)
	s = regexp.MustCompile("(?:[osxz]|[cs]h)$").ReplaceAllString(s, "$0e")
	return s + "s"
}

func Open(s string) string {
	return "{"
}

func Close(s string) string {
	return "}"
}

func Nothing(s string) string {
	return ""
}
