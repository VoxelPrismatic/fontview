package gui

import (
	"fmt"
	"sync"

	"github.com/mappu/miqt/qt6"
)

var (
	tableWidget   *qt6.QTableWidget
	tableScroller *qt6.QScrollBar
	tableMut      sync.Mutex
	tbl_autoSize  = true
	tbl_col_w     = 0
)

func renderGlyphs() {
	if !tableMut.TryLock() {
		return
	}

	rows := tableWidget.RowCount()
	sheetN := tableScroller.Value() - (rows / 3)
	w := tableWidget.ColumnWidth(0)

	curCell := tableWidget.CurrentIndex()
	curR, curC := curCell.Row(), curCell.Column()

	for idx := range rows {
		tableWidget.SetRowHeight(idx, w)
	}

	for idx := rows / 3; idx <= rows/3*2; idx++ {
		item := tableWidget.VerticalHeaderItem(idx)
		text := fmt.Sprintf("%03X_", idx+sheetN)
		if item == nil {
			item = qt6.NewQTableWidgetItem2(text)
			item.SetFont(monoFont)
			tableWidget.SetVerticalHeaderItem(idx, item)
		} else {
			item.SetText(text)
		}

		for col := range 16 {
			char := rune(16*(sheetN+idx) + col)
			renderGlyph(char, idx, col, curR, curC)
		}
	}
	tableMut.Unlock()
}

func renderGlyph(char rune, idx, col, curR, curC int) {
	isCur := col == curC && idx == curR
	cell := tableWidget.CellWidget(idx, col)
	render := makeLabel(char, isCur)
	var label *qt6.QLabel

	if cell == nil {
		label = qt6.NewQLabel2()
		label.SetFont(render.Font)
		label.SetAlignment(qt6.AlignCenter)
		cell = label.QWidget
		tableWidget.SetCellWidget(idx, col, label.QWidget)
	} else {
		label = qt6.UnsafeNewQLabel(cell.Metacast("QLabel"))
	}

	label.SetText(render.Text)
	s, t := label.Font(), render.Font
	if s.Family() != t.Family() || s.PixelSize() != t.PixelSize() {
		label.SetFont(render.Font)
	}
	if label.StyleSheet() != render.Style {
		label.SetStyleSheet(render.Style)
	}

}

func UpdateRealFont() {
	w := tableWidget.ColumnWidth(0)
	px := max(int(float64(w)*0.6), 4)
	setFont := qt6.NewQFont2(fontBox.CurrentFont().Family())
	setFont.SetPixelSize(px)
	rawFont := qt6.QRawFont_FromFont(setFont)
	fontPair = FontPair{rawFont, setFont}
	renderGlyphs()
	go func() {
		var maxRune rune
		fam := fontPair.Raw.FamilyName()
		maxGlyphMut.Lock()
		done := maxGlyphSuccess[fam]
		maxGlyphMut.Unlock()

		if !done {
			maxRune = maxGlyph()
		} else {
			maxGlyphMut.Lock()
			maxRune = rune(maxGlyphCache[fam])
			maxGlyphMut.Unlock()
		}
		if maxRune <= 0 {
			return
		}
		maxRune += maxRune % 16

		tableScroller.SetMaximum(int(maxRune / 16))
	}()
}

func tbl_OnKeyEvt(super func(evt *qt6.QKeyEvent), evt *qt6.QKeyEvent) {
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

func resizeGlyphs() {
	excess := 16
	if tbl_autoSize {
		tbl_w := contentSize(tableWidget) - 8 -
			contentSize(tableWidget.VerticalHeader())
		tbl_col_w = max(tbl_w/16, 4)
		excess -= (tbl_w % 16)
	}

	for column := range tableWidget.ColumnCount() {
		if column >= excess {
			tableWidget.SetColumnWidth(column, tbl_col_w+1)
		} else {
			tableWidget.SetColumnWidth(column, tbl_col_w)
		}
	}
	tableWidget.SetRowCount((tableWidget.Size().Height() / tbl_col_w) * 3)
	tableWidget.VerticalScrollBar().SetValue(tableWidget.RowCount() / 3)
	UpdateRealFont()
	renderGlyphs()
}
