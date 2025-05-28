package gui

import (
	"fmt"
	"strings"

	"github.com/mappu/miqt/qt6"
)

func makeStyleSheet(styles map[string]map[string]string) string {
	sb := strings.Builder{}
	for cls, obj := range styles {
		sb.WriteString(fmt.Sprintf("%s {\n", cls))
		for prop, val := range obj {
			sb.WriteString(fmt.Sprintf("\t%s: %s;\n", prop, val))
		}
		sb.WriteString("}\n")
	}
	return sb.String()
}

type Widthable interface {
	Size() *qt6.QSize
	ContentsMargins() *qt6.QMargins
}

func contentSize(obj Widthable) int {
	margins := obj.ContentsMargins()
	sz := obj.Size()
	w := sz.Width() + margins.Left() + margins.Right()
	return w
}
