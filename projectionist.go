package proj

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var ProjectionFile = ".projections.json"

type Props map[string]string
type Projection map[string]Props
type Projections map[string]Projection

func Project(filename string) Props {
	return Props(map[string]string{})
}

var matchSplit *regexp.Regexp
var loneStar *regexp.Regexp

func init() {
	matchSplit = regexp.MustCompile(`\*\*?`)
	loneStar = regexp.MustCompile(`^[^*{}]*\*[^*{}]*$`)
}

func normPatt(patt string) string {
	if loneStar.MatchString(patt) {
		patt = strings.Replace(patt, "*", "**/*", 1)
	}
	return patt
}

func matches(pattern, filename string) (string, bool) {
	pattern = normPatt(pattern)
	comp := matchSplit.Split(pattern, -1)
	if len(comp) == 1 {
		return "", pattern == filename
	}
	if len(comp) != 3 {
		panic(fmt.Sprintf("%s: Should have splet the path in 3: %#v", pattern, comp))
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

var transformFuncs = map[string]func(s string) string{
	"dot":        Dot,
	"underscore": Underscore,
	"backslash":  Backslash,
	"colons":     Colons,
	"hyphenate":  Hyphenate,
	"blank":      Blank,
	"uppercase":  UpperCase,
	"camelcase":  CamelCase,
	"snakecase":  SnakeCase,
	"capitalize": Capitalize,
	"dirname":    Dirname,
	"basename":   Basename,
	"singular":   Singular,
	"plural":     Plural,
	"open":       Open,
	"Close":      Close,
	"nothing":    Nothing,
}
var placeholderPattern *regexp.Regexp

func init() {
	placeholderPattern = regexp.MustCompile("{[^{}]*}")
}

func ExpandPlaceholder(h string, expansions map[string]string) string {
	if len(h) < 2 || h[0] != '{' || h[len(h)-1] != '}' {
		panic(fmt.Sprintf("%s do not look like a placeholder", h))
	}
	transforms := strings.Split(h[1:len(h)-1], "|")
	var value, exists = expansions[transforms[0]]
	if !exists {
		value, exists = expansions["match"]
		if !exists {
			panic("no matches in expensions")
		}
	} else {
		transforms = transforms[1:]
	}

	for _, t := range transforms {
		f, exists := transformFuncs[t]
		if exists {
			value = f(value)
		}
	}

	return value
}

func ExpandPlaceholders(pattern string, expansions map[string]string) (string, error) {
	var b strings.Builder

	start := 0
	for _, m := range placeholderPattern.FindAllStringIndex(pattern, -1) {
		b.WriteString(pattern[start:m[0]])
		b.WriteString(ExpandPlaceholder(pattern[m[0]:m[1]], expansions))
		start = m[1]
	}
	b.WriteString(pattern[start:])
	return b.String(), nil
}

type RawResult struct {
	Value      string
	Expansions map[string]string
}

func pathSplit(p string) []string {
	var comp = []string{}
	p = path.Clean(p)
	for {
		d, f := path.Split(p)
		if f != "" {
			comp = append(comp, f)
		}
		if d == "/" {
			comp = append(comp, d)
			break
		}
		if d == "" {
			break
		}
		p = path.Clean(d)
	}
	rev := make([]string, len(comp))
	for i, c := range comp {
		rev[len(comp)-1-i] = c
	}
	return rev
}

func compPatt(a, b string) int {
	a, b = normPatt(a), normPatt(b)
	aWild := strings.Count(a, "*")
	bWild := strings.Count(b, "*")
	x := aWild - bWild
	if x != 0 {
		return x
	}
	aSlashes := strings.Count(a, "/")
	bSlashes := strings.Count(b, "/")
	x = bSlashes - aSlashes
	if x != 0 {
		return x
	}
	return strings.Compare(b, a)
}

func sorted(m interface{}, f func(a, b string) int, reverse bool) []string {
	mm := reflect.ValueOf(m)
	kks := make([]string, 0, mm.Len())
	sign := 1
	if reverse {
		sign = -1
	}
	it := mm.MapRange()
	for it.Next() {
		k := it.Key().Interface().(string)
		i := sort.Search(len(kks), func(j int) bool {
			return (sign * f(kks[j], k)) <= 0
		})
		kks = kks[:len(kks)+1]
		for j := len(kks) - 1; j > i; j-- {
			kks[j] = kks[j-1]
		}
		kks[i] = k
	}
	return kks
}

func QueryRaw(key, file string, projections Projections) []RawResult {
	var candidates = []RawResult{}

	for _, path := range sorted(projections, compPatt, true) {
		expansions := map[string]string{
			"project": path,
			"file":    file,
		}
		name := ""
		if len(file) >= len(path) && file[:len(path)] == path {
			name = file[len(path):]
		} else {
			continue
		}
		if name[0] == '/' {
			name = name[1:]
		}
		projection := projections[path]
		for _, pattern := range sorted(projection, compPatt, true) {
			props := projection[pattern]
			if match, doesMatch := matches(pattern, name); doesMatch {
				if value, ok := props[key]; ok {
					exp := make(map[string]string)
					for k, v := range expansions {
						exp[k] = v
					}
					exp["match"] = match
					candidates = append(candidates, RawResult{value, exp})
				}
			}
		}
	}

	return candidates
}

func Query(key, file string, projections Projections) [][2]string {
	raw := QueryRaw(key, file, projections)
	var candidates = make([][2]string, 0, len(raw))
	for _, r := range raw {
		value, err := ExpandPlaceholders(r.Value, r.Expansions)
		project, ok := r.Expansions["project"]
		if err == nil && ok {
			rr := [2]string{project, value}
			candidates = append(candidates, rr)
		}
	}
	return candidates
}

func QueryFile(key, file string, projections Projections) []string {
	qqs := Query(key, file, projections)
	rrs := make([]string, 0, len(qqs))
	for _, q := range qqs {
		full := path.Join(q[0], q[1])
		rrs = append(rrs, full)
	}
	return rrs
}

func QueryFileRec(key, file string, rec int, projections Projections) []string {
	depth := 0
	files := []string{}
	currentFiles := []string{file}
	visited := map[string]bool{}

	for {
		if len(currentFiles) == 0 || depth >= rec {
			break
		}
		nextFiles := []string{}
		for _, file := range currentFiles {
			candidates := QueryFile(key, file, projections)
			for _, c := range candidates {
				if !visited[c] {
					visited[c] = true
					files = append(files, c)
					nextFiles = append(nextFiles, c)
				}
			}
		}
		currentFiles = nextFiles
		depth++
	}

	return files
}

func QueryScalar(key, file string, projections Projections) []string {
	qqs := Query(key, file, projections)
	rrs := make([]string, 0, len(qqs))
	for _, c := range qqs {
		rrs = append(rrs, c[1])
	}
	return rrs
}
