package tables

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func RuneToUint(r rune) uint {
	return uint(r&0x7F_FF_FF_FF) + (uint(r) & 0x80_00_00_00)
}

func encode_UTF8(r rune) string {
	bytes := make([]byte, 8)
	_ = utf8.EncodeRune(bytes, r)
	slices.Reverse(bytes)
	for len(bytes) > 1 && bytes[0] == 0 {
		bytes = bytes[1:]
	}
	encoded := make([]string, len(bytes))
	for i, b := range bytes {
		encoded[i] = fmt.Sprintf("0x%02X", uint8(b))
	}

	return strings.Join(encoded, " ")
}

func encode_UTF16(r rune) string {
	rBig, rLittle := utf16.EncodeRune(r)
	if rBig == 0xFFFD && rLittle == 0xFFFD {
		rLittle = r
		rBig = 0
	}
	if rBig > 0 {
		return fmt.Sprintf("0x%04X 0x%04X", RuneToUint(rBig), RuneToUint(rLittle))
	}

	return fmt.Sprintf("0x%04X", RuneToUint(rLittle))
}

func encode_UTF32(r rune) string {
	return fmt.Sprintf("0x%06X", RuneToUint(r))
}

func encode_C_Octal(r rune) string {
	bytes := make([]byte, 8)
	_ = utf8.EncodeRune(bytes, r)
	slices.Reverse(bytes)
	for len(bytes) > 1 && bytes[0] == 0 {
		bytes = bytes[1:]
	}
	encoded := make([]string, len(bytes)+1)
	encoded[0] = ""
	for i, b := range bytes {
		encoded[i+1] = fmt.Sprintf("%o", uint8(b))
	}

	return strings.Join(encoded, "\\")
}

func encode_C_Hex(r rune) string {
	bytes := make([]byte, 8)
	_ = utf8.EncodeRune(bytes, r)
	slices.Reverse(bytes)
	for len(bytes) > 1 && bytes[0] == 0 {
		bytes = bytes[1:]
	}
	encoded := make([]string, len(bytes)+1)
	encoded[0] = ""
	for i, b := range bytes {
		encoded[i+1] = fmt.Sprintf("%x", uint8(b))
	}

	return strings.Join(encoded, "\\x")
}

func encode_XML(r rune) string {
	return fmt.Sprintf("&#%d;", RuneToUint(r))
}

func encode_HTML(r rune) string {
	entites, ok := htmlList[r]
	if !ok || len(entites) == 0 {
		return fmt.Sprintf("&#x%x;", RuneToUint(r))
	}

	return strings.Join(entites, "\n")
}

func encode_JS(r rune) string {
	return fmt.Sprintf("\\u{%04x}", RuneToUint(r))
}

func encode_C_Uni(r rune) string {
	if r > 0xFFFF {
		return fmt.Sprintf("\\U%08x", RuneToUint(r))
	}

	return fmt.Sprintf("\\u%04x", RuneToUint(r))
}

var CodeEncoder = map[string]func(rune) string{
	"UTF-8":       encode_UTF8,
	"UTF-16":      encode_UTF16,
	"UTF-32":      encode_UTF32,
	"\\Octal":     encode_C_Octal,
	"\\Hex":       encode_C_Hex,
	"\\Unicode":   encode_C_Uni,
	"XML Entity":  encode_XML,
	"HTML Entity": encode_HTML,
	"JavaScript":  encode_JS,
}

var CategoryMap = map[string]map[string]string{
	"L": {
		"!": "Letter",
		"u": "Uppercase",
		"l": "Lowercase",
		"t": "Titlecase",
		"m": "Modifier",
		"o": "Other",
	},
	"M": {
		"!": "Mark",
		"n": "Nonspacing",
		"c": "Spacing",
		"e": "Enclosing",
	},
	"N": {
		"!": "Number",
		"d": "Decimal",
		"l": "Letter",
		"o": "Other",
	},
	"P": {
		"!": "Punctuation",
		"c": "Connector",
		"d": "Dash",
		"s": "Open",
		"e": "Close",
		"i": "Initial",
		"f": "Final",
		"o": "Other",
	},
	"S": {
		"!": "Symbol",
		"m": "Math",
		"c": "Currency",
		"k": "Modifier",
		"o": "Other",
	},
	"Z": {
		"!": "Separator",
		"s": "Space",
		"l": "Line",
		"p": "Paragraph",
	},
	"C": {
		"!": "Other",
		"c": "Control",
		"f": "Format",
		"s": "Surrogate",
		"o": "Private Use",
		"n": "Unassigned",
	},
}
