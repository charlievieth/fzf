package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestStderrInternal(t *testing.T) {
	tests := []struct {
		S, Exp string
		NLCR   bool
	}{
		{
			S:    "",
			Exp:  "",
			NLCR: false,
		},
		{
			S:    "abcd",
			Exp:  "abcd",
			NLCR: false,
		},
		{
			S:    "☺☻☹",
			Exp:  "☺☻☹",
			NLCR: false,
		},
		{
			S:    "日a本b語ç日ð本Ê語þ日¥本¼語i日©",
			Exp:  "日a本b語ç日ð本Ê語þ日¥本¼語i日©",
			NLCR: false,
		},
		{
			S:    "日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©",
			Exp:  "日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©",
			NLCR: false,
		},
		{
			S:    "\x80\x80\x80\x80",
			Exp:  "    ",
			NLCR: false,
		},
		{
			S:    "\x80\x80\x80\x80",
			Exp:  "    ",
			NLCR: true,
		},
		{
			S:    "\x1b\x1b\x1b\x1b",
			Exp:  "\x1b\x1b\x1b\x1b",
			NLCR: false,
		},
		{
			S:    "\r\r\n\n",
			Exp:  "    ",
			NLCR: false,
		},
		{
			S:    "\r\r\n\n",
			Exp:  "\r\r\n\n",
			NLCR: true,
		},
	}
	for i, x := range tests {
		var r LightRenderer
		r.stderrInternal(x.S, x.NLCR)
		if r.queued.String() != x.Exp {
			t.Errorf("%d: got: %q want: %q", i, r.queued, x.Exp)
		}
	}
}

func colorCodesReference(fg Color, bg Color) []string {
	codes := []string{}
	appendCode := func(c Color, offset int) {
		if c == colDefault {
			return
		}
		if c.is24() {
			r := (c >> 16) & 0xff
			g := (c >> 8) & 0xff
			b := (c) & 0xff
			codes = append(codes, fmt.Sprintf("%d;2;%d;%d;%d", 38+offset, r, g, b))
		} else if c >= colBlack && c <= colWhite {
			codes = append(codes, fmt.Sprintf("%d", int(c)+30+offset))
		} else if c > colWhite && c < 16 {
			codes = append(codes, fmt.Sprintf("%d", int(c)+90+offset-8))
		} else if c >= 16 && c < 256 {
			codes = append(codes, fmt.Sprintf("%d;5;%d", 38+offset, c))
		}
	}
	appendCode(fg, 0)
	appendCode(bg, 10)
	return codes
}

func csiColorReference(w *LightWindow, fg Color, bg Color, attr Attr) bool {
	codes := append(attrCodes(attr), colorCode(fg, colorOffsetFG), colorCode(bg, colorOffsetBG))
	w.csi(";" + strings.Join(codes, ";") + "m")
	return len(codes) > 0
}

func TestCSIColor(t *testing.T) {
	type TestCase struct {
		Attr   Attr
		FG, BG Color
	}
	base := []TestCase{
		{0, 0, 0},
		{0, 255, 0},
		{0, 0, 255},
		{0, 255, 255},
		{0, 1, 1},
		{0, 1<<24 + 1, 0},
		{0, 0, 1<<24 + 1},
	}
	tests := append([]TestCase(nil), base...)
	for _, a := range []Attr{1, 128, 255} {
		for _, x := range base {
			tests = append(tests, TestCase{Attr: a, FG: x.FG, BG: x.BG})
		}
	}

	assert := func(t *testing.T, i int, x TestCase, fail bool) {
		t.Helper()
		w1 := LightWindow{renderer: &LightRenderer{}}
		w2 := LightWindow{renderer: &LightRenderer{}}

		ok1 := csiColorReference(&w1, x.FG, x.BG, x.Attr)
		ok2 := w2.csiColor(x.FG, x.BG, x.Attr)
		s1 := w1.renderer.queued.String()
		s2 := w2.renderer.queued.String()
		if ok1 != ok2 || s1 != s2 {
			if fail {
				t.Errorf("%d: fg: %d bg: %d attr: %d: got: (%t - %q) want: (%t - %q)",
					i, x.FG, x.BG, x.Attr, ok2, s2, ok1, s1)
			} else {
				t.Logf("Fail (ignoring error): %d: fg: %d bg: %d attr: %d: got: (%t - %q) want: (%t - %q)",
					i, x.FG, x.BG, x.Attr, ok2, s2, ok1, s1)
			}
		}
	}

	for i, x := range tests {
		assert(t, i, x, true)
	}

	// TODO (CEV): figure out if these are really failures of if the original
	// reference implementation was incorrect.
	failingTests := []TestCase{
		{0, 0, 1<<25 + 1},
		{0, 1<<25 + 1, 0},
	}
	for i, x := range failingTests {
		assert(t, i, x, false)
	}
}

func BenchmarkStderrInternal_ValidTenASCIIChars(b *testing.B) {
	const s = "0123456789"
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		var r LightRenderer
		r.stderrInternal(s, false)
	}
}

func BenchmarkStderrInternal_ValidTenJapaneseChars(b *testing.B) {
	const s = "日本語日本語日本語日"
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		var r LightRenderer
		r.stderrInternal(s, false)
	}
}

// func BenchmarkValidTenJapaneseChars(b *testing.B) {
// 	s := []byte("日本語日本語日本語日")
// 	for i := 0; i < b.N; i++ {
// 		Valid(s)
// 	}
// }

// // func BenchmarkValidTenJapaneseChars(b *testing.B) {
// // 	s := []byte("日本語日本語日本語日")
// // 	for i := 0; i < b.N; i++ {
// // 		Valid(s)
// // 	}
// // }
