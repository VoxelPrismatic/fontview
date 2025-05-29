package gui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mappu/miqt/qt6"
)

type FontCache[T any] map[string]map[rune]T

type Render struct {
	Text  string
	Font  *qt6.QFont
	Style string
}

var (
	selectedCache   = FontCache[Render]{}
	labelCache      = FontCache[Render]{}
	fallbackCache   = map[rune]string{}
	supportsCache   = FontCache[bool]{}
	maxGlyphCache   = map[string]rune{}
	maxGlyphSuccess = map[string]bool{}
	maxGlyphMut     sync.Mutex
)

func maxGlyph() rune {
	target := fontPair.Raw
	fam := target.FamilyName()
	last, ok := maxGlyphCache[fam]
	if !ok {
		if len(blocks) == 0 {
			return 0xffff
		}
		last = blocks[len(blocks)-1].End
	}

	for code := last; code >= 0; code-- {
		if fontPair.Raw.SupportsCharacter(uint(code)) {
			maxGlyphMut.Lock()
			maxGlyphSuccess[fam] = true
			fmt.Printf("%s: %d\n", fam, code)
			maxGlyphCache[fam] = code
			maxGlyphMut.Unlock()
			return rune(code)
		}
		if target != fontPair.Raw {
			maxGlyphMut.Lock()
			maxGlyphCache[fam] = code
			maxGlyphMut.Unlock()
			return 0
		}
	}
	return 0
}

func runeSupported(r rune) bool {
	fam := fontPair.Raw.FamilyName()
	cache, ok := supportsCache[fam]
	if !ok || cache == nil {
		cache = map[rune]bool{}
		supportsCache[fam] = cache
	}

	if _, ok := cache[r]; !ok {
		cache[r] = fontPair.Raw.SupportsCharacter(uint(r))
	}

	return cache[r]
}

func makeLabel(r rune, selected bool) Render {
	fam := fontPair.Real.Family()
	targetCache := labelCache
	if selected {
		targetCache = selectedCache
	}

	cache, ok := targetCache[fam]
	if !ok || cache == nil {
		cache = map[rune]Render{}
		targetCache[fam] = cache
	}

	if ret, ok := cache[r]; ok {
		return ret
	}

	ret := Render{}
	ret.Text = string(r)

	if runeSupported(r) {
		ret.Font = fontPair.Real
		ret.Style = ""
	} else {
		ret.Text = runeFallback(r)
		ret.Font = monoFont
		ret.Style = "background-color: " + sakurapine.Hl.Low + ";"
		ret.Style += "color: " + sakurapine.Text.Muted + ";"
	}

	if selected {
		ret.Style = "color: " + sakurapine.Layer.Base + ";"
		ret.Style += "font-weight: bold;"
	}

	cache[r] = ret
	return ret
}

func runeFallback(r rune) string {
	if ret, ok := fallbackCache[r]; ok {
		return ret
	}

	st := fmt.Sprintf("%04X", int(r))
	if len(st) >= 8 {
		st := strings.Repeat("0", len(st)%4) + st
		parts := []string{}
		for i := 0; i < len(st); i += 4 {
			parts = append(parts, st[i:i+4])
		}
		st = strings.Join(parts, "<br>")
	} else {
		st = strings.Repeat("0", len(st)%2) + st
		st = st[:len(st)/2] + "<br>" + st[len(st)/2:]
	}
	fallbackCache[r] = st
	return st
}
