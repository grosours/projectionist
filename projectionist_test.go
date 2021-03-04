package proj

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func preparePath(t *testing.T, path string) {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0777); err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func prepareDir(t *testing.T, paths []string) {
	for _, path := range paths {
		preparePath(t, path)
	}
}

var detectTests = []struct {
	tree        []string
	path        string
	projections []string
}{
	{
		[]string{},
		"",
		[]string{},
	},
	{
		[]string{".projections.json"},
		"",
		[]string{".projections.json"},
	},
	{
		[]string{"a/.projections.json"},
		"",
		[]string{},
	},
	{
		[]string{"a/.projections.json", "a/dummy"},
		"a/dummy",
		[]string{"a/.projections.json"},
	},
	{
		[]string{".projections.json", "a/.projections.json", "a/dummy"},
		"a/dummy",
		[]string{"a/.projections.json", ".projections.json"},
	},
}

func TestDetect(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testProjectionist")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	for _, tt := range detectTests {
		temp, err := ioutil.TempDir(testDir, "tt")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(temp); err != nil {
			t.Fatal(err)
		}
		prepareDir(t, tt.tree)
		var fullpaths = []string{}
		for _, relpath := range tt.projections {
			fullpaths = append(fullpaths, filepath.Join(temp, relpath))
		}
		detected, err := Detect(tt.path)
		if len(tt.projections) != len(detected) {
			t.Errorf("Expected %d files, but got %d", len(tt.projections), len(detected))
		}
		for f := range detected {
			dd, pp := detected[f], tt.projections[f]
			d, err := os.Stat(dd)
			if err != nil {
				t.Fatal(dd, err)
			}
			p, err := os.Stat(pp)
			if err != nil {
				t.Fatal(pp, err)
			}
			if !os.SameFile(d, p) {
				t.Errorf("%s != %s", dd, pp)
			}
		}
	}
}

var matchesTest = []struct {
	pattern, file, sub string
	match              bool
}{
	{"*.c", "bla.c", "bla", true},
	{"**/*.c", "bla.c", "bla", true},
	{"**/*.c", "coucou/bla.c", "coucou/bla", true},
	{"**/*.a", "bla.c", "", false},
	{"coucou/**/test_*.c", "coucou/bla/test_bli.c", "bla/bli", true},
	{"coucou/**test_*.c", "coucou/bla/test_bli.c", "bla/bli", true},
	{"b", "b", "", true},
}

func TestMatches(t *testing.T) {
	for _, tt := range matchesTest {
		sub, match := matches(tt.pattern, tt.file)
		t.Logf("%q | %q -> %q", tt.pattern, tt.file, tt.sub)
		assert.Equal(t, tt.sub, sub)
		assert.Equal(t, tt.match, match)
	}
}

var dotTests = []struct {
	input, output string
}{
	{"a/b", "a.b"},
	{"a/b/c", "a.b.c"},
}

func TestDot(t *testing.T) {
	for _, tt := range dotTests {
		out := Dot(tt.input)
		assert.Equal(t, tt.output, out)
	}
}

var camelCaseTests = []struct {
	in, out string
}{
	{"foo_bar/baz_quux", "fooBar/bazQuux"},
	{"foo_bar/baz_quux_", "fooBar/bazQuux"},
	{"_foo_bar/_baz_quux_", "FooBar/BazQuux"},
}

func TestCamelCase(t *testing.T) {
	for _, tt := range camelCaseTests {
		out := CamelCase(tt.in)
		assert.Equal(t, tt.out, out)
	}
}

var snakeCaseTests = []struct {
	in, out string
}{
	{"fooBar/bazQuux", "foo_bar/baz_quux"},
	{"fooBar/bazQuux", "foo_bar/baz_quux"},
	{"FooBar/BazQuux", "_foo_bar/_baz_quux"},
}

func TestSnakeCase(t *testing.T) {
	for _, tt := range snakeCaseTests {
		out := SnakeCase(tt.in)
		assert.Equal(t, tt.out, out)
	}
}

var singularTests = []struct {
	in, out string
}{
	{"jiffies", "jiffy"},
	{"movies", "movie"},
}

func TestSingular(t *testing.T) {
	for _, tt := range singularTests {
		out := Singular(tt.in)
		assert.Equal(t, tt.out, out)
	}
}

var pluralTests = []struct {
	in, out string
}{
	{"jiffy", "jiffies"},
	{"movie", "movies"},
}

func TestPlural(t *testing.T) {
	for _, tt := range pluralTests {
		out := Plural(tt.in)
		assert.Equal(t, tt.out, out)
	}
}

var expandPlaceholderTests = []struct {
	in         string
	expensions map[string]string
	out        string
}{
	{"{}", map[string]string{"match": "a/b"}, "a/b"},
	{"{file}", map[string]string{"file": "a/b"}, "a/b"},
	{"{dot|underscore}", map[string]string{"match": "a/b"}, "a.b"},
	{"{dot|uppercase}", map[string]string{"match": "a/b"}, "A.B"},
	{"{dirname}", map[string]string{"match": "a/b"}, "a"},
}

func TestExpandPlaceholder(t *testing.T) {
	for _, tt := range expandPlaceholderTests {
		out := ExpandPlaceholder(tt.in, tt.expensions)
		assert.Equal(t, tt.out, out)
	}
}

func TestExpandPlaceholderPanics(t *testing.T) {
	assert.Panics(t, func() { ExpandPlaceholder("", nil) })
	assert.Panics(t, func() { ExpandPlaceholder("{}", map[string]string{}) })
}

var expandPlaceholdersTests = []struct {
	in, out    string
	expansions map[string]string
}{
	{"{dirname|dot}/{basename}", "a.b/c", map[string]string{"match": "a/b/c"}},
	{"prefix-{dirname|dot}/{basename}", "prefix-a.b/c", map[string]string{"match": "a/b/c"}},
	{"{dirname|dot}/{basename}-suffix", "a.b/c-suffix", map[string]string{"match": "a/b/c"}},
}

func TestExpandPlaceholders(t *testing.T) {
	for _, tt := range expandPlaceholdersTests {
		out, err := ExpandPlaceholders(tt.in, tt.expansions)
		assert.Nil(t, err, "Should not have returned an error")
		assert.Equal(t, tt.out, out)
	}
}

func TestQueryRaw(t *testing.T) {
	var projections = Projections(map[string]Projection{
		"/": Projection(map[string]Props{
			"*": Props(map[string]string{
				"key1": "root1",
				"key2": "root2",
			}),
		}),
		"/a": Projection(map[string]Props{
			"*": Props(map[string]string{
				"key1": "a1",
			}),
			"b": Props(map[string]string{
				"key1": "b1",
			}),
		}),
	})

	for _, tt := range []struct {
		key, file   string
		projections Projections
		values      []string
	}{
		{"key1", "/a/c", projections, []string{"a1", "root1"}},
		{"key1", "/a/b", projections, []string{"b1", "a1", "root1"}},
		{"key2", "/a/d", projections, []string{"root2"}},
		{"key2", "/b/d", projections, []string{"root2"}},
		{"key3", "/a/c", projections, []string{}},
	} {
		t.Log(tt)
		results := QueryRaw(tt.key, tt.file, tt.projections)
		values := make([]string, len(results))
		for i, r := range results {
			values[i] = r.Value
		}
		assert.Equal(t, tt.values, values)
	}
}

func TestCompPatt(t *testing.T) {
	for _, tt := range []struct {
		a, b  string
		count int
	}{
		{"a", "a", 0},
		{"a/**/*.c", "b/**/*.c", 1},
		{"a/**/*.c", "b/**/c/*.d", 1},
		{"a", "a/b", 1},
		{"*", "a", 1},
		{"a/*", "**/a", 1},
		{"a/**/*.c", "a", 1},
		{"a/**/*.c", "a/**/b", 1},
		{"a/**/*.c", "a/*.c", 0},
	} {
		t.Log(tt)
		count := compPatt(tt.a, tt.b)
		if tt.count == 0 {
			assert.Equal(t, tt.count, count)
		} else {
			assert.Greater(t, tt.count*count, 0, "Should be the same signs")
		}
	}
}

func TestPathSplit(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out []string
	}{
		{"a", []string{"a"}},
		{"a/b", []string{"a", "b"}},
		{"/", []string{"/"}},
		{"/*/a", []string{"/", "*", "a"}},
		{"a/", []string{"a"}},
	} {
		out := pathSplit(tt.in)
		t.Log(tt)
		assert.Equal(t, tt.out, out)
	}
}

func TestQuery(t *testing.T) {
	var projections = Projections(map[string]Projection{
		"/": Projection(map[string]Props{
			"*": Props(map[string]string{
				"key1": "{}",
			}),
			"b": Props(map[string]string{
				"key1": "b1",
			}),
		}),
	})

	for _, tt := range []struct {
		key, file string
		out       [][2]string
	}{
		{"key1", "/a", [][2]string{{"/", "a"}}},
		{"key1", "/b", [][2]string{{"/", "b1"}, {"/", "b"}}},
	} {
		t.Log(tt)
		out := Query(tt.key, tt.file, projections)
		assert.Equal(t, tt.out, out)
	}
}

func TestQueryFile(t *testing.T) {
	var projections = Projections(map[string]Projection{
		"/": Projection(map[string]Props{
			"*": Props(map[string]string{
				"key1": "{}",
			}),
			"b": Props(map[string]string{
				"key1": "b1",
			}),
		}),
	})

	for _, tt := range []struct {
		key, file string
		out       []string
	}{
		{"key1", "/a", []string{"/a"}},
		{"key1", "/b", []string{"/b1", "/b"}},
	} {
		t.Log(tt)
		out := QueryFile(tt.key, tt.file, projections)
		assert.Equal(t, tt.out, out)
	}
}

func TestQueryFileRec(t *testing.T) {
	var projections = Projections(map[string]Projection{
		"/": Projection(map[string]Props{
			"*": Props(map[string]string{
				"key1": "{}",
			}),
			"b": Props(map[string]string{
				"key1": "b1",
			}),
			"b1": Props(map[string]string{
				"key1": "b2",
			}),
			"b2": Props(map[string]string{
				"key1": "b3",
			}),
		}),
	})

	for _, tt := range []struct {
		key, file string
		rec       int
		out       []string
	}{
		{"key1", "/a", 10, []string{"/a"}},
		{"key1", "/b", 10, []string{"/b1", "/b", "/b2", "/b3"}},
		{"key1", "/b", 1, []string{"/b1", "/b"}},
	} {
		t.Log(tt)
		out := QueryFileRec(tt.key, tt.file, tt.rec, projections)
		assert.Equal(t, tt.out, out)
	}
}
