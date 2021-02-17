package fzf

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/junegunn/fzf/src/tui"
)

// The following regular expression will include not all but most of the
// frequently used ANSI sequences. This regex is used as a reference for
// testing nextAnsiEscapeSequence().
//
// References:
// 	- https://github.com/gnachman/iTerm2
// 	- http://ascii-table.com/ansi-escape-sequences.php
// 	- http://ascii-table.com/ansi-escape-sequences-vt-100.php
// 	- http://tldp.org/HOWTO/Bash-Prompt-HOWTO/x405.html
// 	- https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
var ansiRegexRefence = regexp.MustCompile("(?:\x1b[\\[()][0-9;]*[a-zA-Z@]|\x1b][0-9];[[:print:]]+(?:\x1b\\\\|\x07)|\x1b.|[\x0e\x0f]|.\x08)")

var ansiBenchmarkStrings = [...]string{
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgithub.com/\x1b[0m\x1b[38;5;81mdocker/\x1b[0m\x1b[38;5;81mdistribution@v2.7.1+incompatible/\x1b[0m\x1b[38;5;81mmanifest/\x1b[0m\x1b[38;5;81mschema1/\x1b[0m\x1b[38;5;48mconfig_builder.go\x1b[0m",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgithub.com/\x1b[0m\x1b[38;5;81mm3db/\x1b[0m\x1b[38;5;81mm3x@v0.0.0-20190408051622-ebf3c7b94afd/\x1b[0m\x1b[38;5;81mconfig/\x1b[0m\x1b[38;5;81mtestdata\x1b[0m",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgolang.org/\x1b[0m\x1b[38;5;81mx/\x1b[0m\x1b[38;5;81mtools@v0.1.1-0.20210122170814-cf1022a4b077/\x1b[0m\x1b[38;5;81minternal/\x1b[0m\x1b[38;5;81mlsp/\x1b[0m\x1b[38;5;81mtestdata/\x1b[0m\x1b[38;5;81mtesty/\x1b[0m\x1b[38;5;48mtesty_test.go\x1b[0m",
	"src/github.com/prometheus/prometheus/web/ui/react-app/src/pages/targets/target.ts",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgithub.com/\x1b[0m\x1b[38;5;81mcockroachdb/\x1b[0m\x1b[38;5;81mcockroach@v20.1.9+incompatible/\x1b[0m\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81msql/\x1b[0m\x1b[38;5;81mrow/\x1b[0m\x1b[38;5;48mfk_existence_base.go\x1b[0m",
	"pkg/mod/github.com/go-delve/delve@v1.5.2-0.20201221185609-e7558c5bc5a3/_fixtures/issue683.go",
	"src/github.com/m3db/m3/src/query/graphite/graphite/glob.go",
	"pkg/mod/github.com/lib/pq@v1.8.0/connector_test.go",
	"pkg/mod/google.golang.org/api@v0.36.0/remotebuildexecution/v2",
	"pkg/mod/golang.org/x/tools@v0.1.1-0.20210122170814-cf1022a4b077/internal/lsp/testdata/errors",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgolang.org/\x1b[0m\x1b[38;5;81mx/\x1b[0m\x1b[38;5;81mtools@v0.0.0-20201030010431-2feb2bb1ff51/\x1b[0m\x1b[38;5;81mcmd/\x1b[0m\x1b[38;5;81mguru/\x1b[0m\x1b[38;5;81mtestdata/\x1b[0m\x1b[38;5;81msrc/\x1b[0m\x1b[38;5;81mcalls-json\x1b[0m",
	"pkg/mod/golang.org/x/tools@v0.1.1-0.20210204180613-842a9283d6c6/refactor",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mhonnef.co/\x1b[0m\x1b[38;5;81mgo/\x1b[0m\x1b[38;5;81mtools@v0.0.1-2020.1.0.20201124073330-56b7c78ddcd8/\x1b[0m\x1b[38;5;81mstaticcheck/\x1b[0m\x1b[38;5;81mtestdata/\x1b[0m\x1b[38;5;81msrc/\x1b[0m\x1b[38;5;81mCheckUntrappableSignal\x1b[0m",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgolang.org/\x1b[0m\x1b[38;5;81mx/\x1b[0m\x1b[38;5;81mtools@v0.1.1-0.20210122203318-2972602ec4f0/\x1b[0m\x1b[38;5;81mgo/\x1b[0m\x1b[38;5;81minternal/\x1b[0m\x1b[38;5;81mgcimporter/\x1b[0m\x1b[38;5;48miexport_test.go\x1b[0m",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgolang.org/\x1b[0m\x1b[38;5;81mx/\x1b[0m\x1b[38;5;81mtools@v0.0.0-20210108195828-e2f9c7f1fc8e/\x1b[0m\x1b[38;5;81mgo/\x1b[0m\x1b[38;5;81manalysis/\x1b[0m\x1b[38;5;81mpasses/\x1b[0m\x1b[38;5;81mifaceassert/\x1b[0m\x1b[38;5;81mtestdata/\x1b[0m\x1b[38;5;81msrc\x1b[0m",
	"src/github.com/cockroachdb/cockroach/pkg/security/securitytest/test_certs/client-tenant.10.crt",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgolang.org/\x1b[0m\x1b[38;5;81mx/\x1b[0m\x1b[38;5;81mtools@v0.0.0-20201130220005-fd5f29369093/\x1b[0m\x1b[38;5;81mgo/\x1b[0m\x1b[38;5;81manalysis/\x1b[0m\x1b[38;5;81mpasses/\x1b[0m\x1b[38;5;81mfieldalignment\x1b[0m",
	"\x1b[38;5;81mpkg/\x1b[0m\x1b[38;5;81mmod/\x1b[0m\x1b[38;5;81mgithub.com/\x1b[0m\x1b[38;5;81mdocker/\x1b[0m\x1b[38;5;81mdocker@v17.12.0-ce-rc1.0.20190115172544-0dc531243dd3+incompatible/\x1b[0m\x1b[38;5;81mdaemon/\x1b[0m\x1b[38;5;81mdiscovery/\x1b[0m\x1b[38;5;48mdiscovery_test.go\x1b[0m",
	"pkg/mod/github.com/golangci/golangci-lint@v1.36.1-0.20210208193152-3ef13a8028c1/test/testdata/skipdirs/skip_me/nested/with_issue.go",
	"pkg/mod/golang.org/x/tools@v0.0.0-20180824175216-6c1c5e93cdc1/cmd/guru/testdata/src/peers-json/main.golden",
}

func testParserReference(t testing.TB, str string) {
	t.Helper()

	toSlice := func(start, end int) []int {
		if start == -1 {
			return nil
		}
		return []int{start, end}
	}

	s := str
	for i := 0; ; i++ {
		got := toSlice(nextAnsiEscapeSequence(s))
		exp := ansiRegexRefence.FindStringIndex(s)

		equal := len(got) == len(exp)
		if equal {
			for i := 0; i < len(got); i++ {
				if got[i] != exp[i] {
					equal = false
					break
				}
			}
		}
		if !equal {
			var exps, gots []rune
			if len(got) == 2 {
				gots = []rune(s[got[0]:got[1]])
			}
			if len(exp) == 2 {
				exps = []rune(s[exp[0]:exp[1]])
			}
			t.Errorf("%d: %q: got: %v (%q) want: %v (%q)", i, s, got, gots, exp, exps)
			return
		}
		if len(exp) == 0 {
			return
		}
		s = s[exp[1]:]
	}
}

func TestNextAnsiEscapeSequence(t *testing.T) {
	testStrs := []string{
		"\x1b[0mhello world",
		"\x1b[1mhello world",
		"椙\x1b[1m椙",
		"椙\x1b[1椙m椙",
		"\x1b[1mhello \x1b[mw\x1b7o\x1b8r\x1b(Bl\x1b[2@d",
		"\x1b[1mhello \x1b[Kworld",
		"hello \x1b[34;45;1mworld",
		"hello \x1b[34;45;1mwor\x1b[34;45;1mld",
		"hello \x1b[34;45;1mwor\x1b[0mld",
		"hello \x1b[34;48;5;233;1mwo\x1b[38;5;161mr\x1b[0ml\x1b[38;5;161md",
		"hello \x1b[38;5;38;48;5;48;1mwor\x1b[38;5;48;48;5;38ml\x1b[0md",
		"hello \x1b[32;1mworld",
		"hello world",
		"hello \x1b[0;38;5;200;48;5;100mworld",
		"\x1b椙",
		"椙\x08",
		"",
	}
	testStrs = append(testStrs, ansiBenchmarkStrings[:]...)

	for _, s := range testStrs {
		testParserReference(t, s)
	}
}

func TestNextAnsiEscapeSequence_Fuzz_Modified(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("short test")
	}

	testStrs := []string{
		"\x1b[0mhello world",
		"\x1b[1mhello world",
		"椙\x1b[1m椙",
		"椙\x1b[1椙m椙",
		"\x1b[1mhello \x1b[mw\x1b7o\x1b8r\x1b(Bl\x1b[2@d",
		"\x1b[1mhello \x1b[Kworld",
		"hello \x1b[34;45;1mworld",
		"hello \x1b[34;45;1mwor\x1b[34;45;1mld",
		"hello \x1b[34;45;1mwor\x1b[0mld",
		"hello \x1b[34;48;5;233;1mwo\x1b[38;5;161mr\x1b[0ml\x1b[38;5;161md",
		"hello \x1b[38;5;38;48;5;48;1mwor\x1b[38;5;48;48;5;38ml\x1b[0md",
		"hello \x1b[32;1mworld",
		"hello world",
		"hello \x1b[0;38;5;200;48;5;100mworld",
	}
	testStrs = append(testStrs, ansiBenchmarkStrings[:]...)

	replacementBytes := [...]rune{'\x0e', '\x0f', '\x1b', '\x08'}

	modifyString := func(s string, rr *rand.Rand) string {
		n := rr.Intn(len(s))
		b := []rune(s)
		for ; n >= 0 && len(b) != 0; n-- {
			i := rr.Intn(len(b))
			switch x := rr.Intn(4); x {
			case 0:
				b = append(b[:i], b[i+1:]...)
			case 1:
				j := rr.Intn(len(replacementBytes) - 1)
				b[i] = replacementBytes[j]
			case 2:
				x := rune(rr.Intn(utf8.MaxRune))
				for !utf8.ValidRune(x) {
					x = rune(rr.Intn(utf8.MaxRune))
				}
				b[i] = x
			case 3:
				b[i] = rune(rr.Intn(utf8.MaxRune)) // potentially invalid
			default:
				t.Fatalf("unsupported value: %d", x)
			}
		}
		return string(b)
	}

	rr := rand.New(rand.NewSource(1))
	for _, s := range testStrs {
		for i := 1_000; i >= 0; i-- {
			testParserReference(t, modifyString(s, rr))
		}
	}
}

func TestNextAnsiEscapeSequence_Fuzz_Random(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("short test")
	}

	randomString := func(rr *rand.Rand) string {
		numChars := rand.Intn(50)
		codePoints := make([]rune, numChars)
		for i := 0; i < len(codePoints); i++ {
			var r rune
			for n := 0; n < 1000; n++ {
				r = rune(rr.Intn(utf8.MaxRune))
				// Allow 10% of runes to be invalid
				if utf8.ValidRune(r) || rr.Float64() < 0.10 {
					break
				}
			}
			codePoints[i] = r
		}
		return string(codePoints)
	}

	rr := rand.New(rand.NewSource(1))
	for i := 0; i < 100_000; i++ {
		testParserReference(t, randomString(rr))
	}
}

func TestExtractColor(t *testing.T) {
	assert := func(offset ansiOffset, b int32, e int32, fg tui.Color, bg tui.Color, bold bool) {
		var attr tui.Attr
		if bold {
			attr = tui.Bold
		}
		if offset.offset[0] != b || offset.offset[1] != e ||
			offset.color.fg != fg || offset.color.bg != bg || offset.color.attr != attr {
			t.Error(offset, b, e, fg, bg, attr)
		}
	}

	src := "hello world"
	var state *ansiState
	clean := "\x1b[0m"
	check := func(assertion func(ansiOffsets *[]ansiOffset, state *ansiState)) {
		output, ansiOffsets, newState := extractColor(src, state, nil)
		state = newState
		if output != "hello world" {
			t.Errorf("Invalid output: %s %v", output, []rune(output))
		}
		fmt.Println(src, ansiOffsets, clean)
		assertion(ansiOffsets, state)
	}

	check(func(offsets *[]ansiOffset, state *ansiState) {
		if offsets != nil {
			t.Fail()
		}
	})

	state = nil
	src = "\x1b[0mhello world"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if offsets != nil {
			t.Fail()
		}
	})

	state = nil
	src = "\x1b[1mhello world"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		assert((*offsets)[0], 0, 11, -1, -1, true)
	})

	state = nil
	src = "\x1b[1mhello \x1b[mw\x1b7o\x1b8r\x1b(Bl\x1b[2@d"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		assert((*offsets)[0], 0, 6, -1, -1, true)
	})

	state = nil
	src = "\x1b[1mhello \x1b[Kworld"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		assert((*offsets)[0], 0, 11, -1, -1, true)
	})

	state = nil
	src = "hello \x1b[34;45;1mworld"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		assert((*offsets)[0], 6, 11, 4, 5, true)
	})

	state = nil
	src = "hello \x1b[34;45;1mwor\x1b[34;45;1mld"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		assert((*offsets)[0], 6, 11, 4, 5, true)
	})

	state = nil
	src = "hello \x1b[34;45;1mwor\x1b[0mld"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		assert((*offsets)[0], 6, 9, 4, 5, true)
	})

	state = nil
	src = "hello \x1b[34;48;5;233;1mwo\x1b[38;5;161mr\x1b[0ml\x1b[38;5;161md"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 3 {
			t.Fail()
		}
		assert((*offsets)[0], 6, 8, 4, 233, true)
		assert((*offsets)[1], 8, 9, 161, 233, true)
		assert((*offsets)[2], 10, 11, 161, -1, false)
	})

	// {38,48};5;{38,48}
	state = nil
	src = "hello \x1b[38;5;38;48;5;48;1mwor\x1b[38;5;48;48;5;38ml\x1b[0md"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 2 {
			t.Fail()
		}
		assert((*offsets)[0], 6, 9, 38, 48, true)
		assert((*offsets)[1], 9, 10, 48, 38, true)
	})

	src = "hello \x1b[32;1mworld"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		if state.fg != 2 || state.bg != -1 || state.attr == 0 {
			t.Fail()
		}
		assert((*offsets)[0], 6, 11, 2, -1, true)
	})

	src = "hello world"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 1 {
			t.Fail()
		}
		if state.fg != 2 || state.bg != -1 || state.attr == 0 {
			t.Fail()
		}
		assert((*offsets)[0], 0, 11, 2, -1, true)
	})

	src = "hello \x1b[0;38;5;200;48;5;100mworld"
	check(func(offsets *[]ansiOffset, state *ansiState) {
		if len(*offsets) != 2 {
			t.Fail()
		}
		if state.fg != 200 || state.bg != 100 || state.attr > 0 {
			t.Fail()
		}
		assert((*offsets)[0], 0, 6, 2, -1, true)
		assert((*offsets)[1], 6, 11, 200, 100, false)
	})
}

func TestAnsiCodeStringConversion(t *testing.T) {
	assert := func(code string, prevState *ansiState, expected string) {
		state := interpretCode(code, prevState)
		if expected != state.ToString() {
			t.Errorf("expected: %s, actual: %s",
				strings.Replace(expected, "\x1b[", "\\x1b[", -1),
				strings.Replace(state.ToString(), "\x1b[", "\\x1b[", -1))
		}
	}
	assert("\x1b[m", nil, "")
	assert("\x1b[m", &ansiState{attr: tui.Blink, lbg: -1}, "")

	assert("\x1b[31m", nil, "\x1b[31;49m")
	assert("\x1b[41m", nil, "\x1b[39;41m")

	assert("\x1b[92m", nil, "\x1b[92;49m")
	assert("\x1b[102m", nil, "\x1b[39;102m")

	assert("\x1b[31m", &ansiState{fg: 4, bg: 4, lbg: -1}, "\x1b[31;44m")
	assert("\x1b[1;2;31m", &ansiState{fg: 2, bg: -1, attr: tui.Reverse, lbg: -1}, "\x1b[1;2;7;31;49m")
	assert("\x1b[38;5;100;48;5;200m", nil, "\x1b[38;5;100;48;5;200m")
	assert("\x1b[48;5;100;38;5;200m", nil, "\x1b[38;5;200;48;5;100m")
	assert("\x1b[48;5;100;38;2;10;20;30;1m", nil, "\x1b[1;38;2;10;20;30;48;5;100m")
	assert("\x1b[48;5;100;38;2;10;20;30;7m",
		&ansiState{attr: tui.Dim | tui.Italic, fg: 1, bg: 1},
		"\x1b[2;3;7;38;2;10;20;30;48;5;100m")
}

func BenchmarkNextAnsiEscapeSequence(b *testing.B) {
	n := 0
	for i := 0; i < len(ansiBenchmarkStrings); i++ {
		n += len(ansiBenchmarkStrings[i])
	}
	b.SetBytes(int64(float64(n) / float64(len(ansiBenchmarkStrings))))

	for i := 0; i < b.N; i++ {
		s := ansiBenchmarkStrings[i%len(ansiBenchmarkStrings)]
		for {
			_, o := nextAnsiEscapeSequence(s)
			if o == -1 {
				break
			}
			s = s[o:]
		}
	}
}

// Baseline test to compare the speed of nextAnsiEscapeSequence() to the
// previously used regex based implementation.
func BenchmarkNextAnsiEscapeSequence_Regex(b *testing.B) {
	n := 0
	for i := 0; i < len(ansiBenchmarkStrings); i++ {
		n += len(ansiBenchmarkStrings[i])
	}
	b.SetBytes(int64(float64(n) / float64(len(ansiBenchmarkStrings))))

	for i := 0; i < b.N; i++ {
		s := ansiBenchmarkStrings[i%len(ansiBenchmarkStrings)]
		for {
			a := ansiRegexRefence.FindStringIndex(s)
			if len(a) == 0 {
				break
			}
			s = s[a[1]:]
		}
	}
}

func BenchmarkExtractColor(b *testing.B) {
	n := 0
	for i := 0; i < len(ansiBenchmarkStrings); i++ {
		n += len(ansiBenchmarkStrings[i])
	}
	b.SetBytes(int64(float64(n) / float64(len(ansiBenchmarkStrings))))

	for i := 0; i < b.N; i++ {
		var state *ansiState
		extractColor(ansiBenchmarkStrings[i%len(ansiBenchmarkStrings)], state, nil)
	}
}
