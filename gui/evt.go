package gui

import (
	"fmt"

	"github.com/mappu/miqt/qt6"
)

var (
	tbl_ignoreEvt = false
)

func tbl_KeyEvt(super func(evt *qt6.QKeyEvent), evt *qt6.QKeyEvent) {
	v := tableScroller.Value()
	top := tableWidget.RowCount() / 3

	dy := top
	mods := evt.Modifiers()
	if mods&qt6.ControlModifier > 0 {
		dy = 0x100
	}
	switch qt6.Key(evt.Key()) {
	case qt6.Key_PageDown:
		tableScroller.SetValue(v + dy)
	case qt6.Key_PageUp:
		tableScroller.SetValue(v - dy)
	case qt6.Key_Down:
		if tableWidget.CurrentRow() >= (top-1)*2 {
			tableScroller.SetValue(v + 1)
			evt.Ignore()
		} else {
			super(evt)
		}
	case qt6.Key_Up:
		if tableWidget.CurrentRow() <= top {
			tableScroller.SetValue(v - 1)
			evt.Ignore()
		} else {
			super(evt)
		}
	case qt6.Key_Equal:
		if mods&qt6.ControlModifier > 0 {
			tbl_autoSize = false
			tbl_col_w += 1
			resizeGlyphs()
		}
	case qt6.Key_Minus:
		if mods&qt6.ControlModifier > 0 {
			tbl_autoSize = false
			tbl_col_w = max(tbl_col_w-1, 4)
			resizeGlyphs()
		}
	case qt6.Key_0:
		if mods&qt6.ControlModifier > 0 {
			tbl_autoSize = true
			resizeGlyphs()
		}

	default:
		super(evt)
	}
}

func tbl_ScrollEvt(super func(dx int, dy int), dx int, dy int) {
	if tbl_ignoreEvt {
		tbl_ignoreEvt = false
		return
	}

	tbl_ignoreEvt = true
	sheetMax := tableScroller.Maximum()
	sheetMin := tableScroller.Minimum()
	sheetNew := max(min(tableScroller.Value()-dy, sheetMax), sheetMin)
	tableScroller.SetValue(sheetNew)
	top := tableWidget.RowCount() / 3
	tableWidget.VerticalScrollBar().SetValue(top)
	super(dx, 0)
}

func tbl_CellChanged(_, _, _, _ int) {
	renderGlyphs()
}

func tbl_ScrollChanged(value int) {
	renderGlyphs()
}

func tbl_ResizeEvt(_ func(_ *qt6.QResizeEvent), evt *qt6.QResizeEvent) {
	labelCache = FontCache[Render]{}
	selectedCache = FontCache[Render]{}
	resizeGlyphs()
}

func btn_DirEvt(prev *[]string, fwd *[]string) {
	l := len(*prev)
	if l == 0 {
		return
	}

	target := (*prev)[l-1]
	*prev = (*prev)[:l-1]
	*fwd = append(*fwd, fmt.Sprint(curNode.Point))
	historyPush = false
	onLink(target)
}

func btn_FwdEvt() {
	btn_DirEvt(&fwdStack, &backStack)
}

func btn_BackEvt() {
	btn_DirEvt(&backStack, &fwdStack)
}
