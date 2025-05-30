package gui

import (
	"fmt"
	"strconv"
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

var fwdStack = []string{}
var backStack = []string{}
var historyPush = true

func onLink(link string) {
	point, err := strconv.Atoi(link)
	if err != nil {
		panic(err)
	}
	if historyPush {
		fwdStack = []string{}
		backStack = append(backStack, fmt.Sprint(curNode.Point))
	} else {
		historyPush = true
	}
	fmt.Println("Back:", backStack)
	fmt.Println("Fwd: ", fwdStack)
	btnBack.SetDisabled(len(backStack) == 0)
	btnFwd.SetDisabled(len(fwdStack) == 0)
	third := tableWidget.RowCount() / 3
	off := tableWidget.CurrentRow() - third
	row := (point-point%16)/16 - off
	tableScroller.SetValue(row)
	tableWidget.SetCurrentCell(off+third, point%16)
}
