package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"fontview/tables"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
)

var (
	blocks    []tables.Block
	blocksMut sync.Mutex
	names     map[string]tables.Node
	namesMut  sync.Mutex
)

var (
	fontBox       *qt6.QFontComboBox
	tableWidget   *qt6.QTableWidget
	tableScroller *qt6.QScrollBar
	tableMut      sync.Mutex
	window        *qt6.QMainWindow
	msg           *qt6.QProgressDialog
)

func readTables() {
	if msg != nil || blocks != nil || names != nil {
		return
	}

	mainthread.Wait(func() {
		msg = qt6.NewQProgressDialog(nil)
		msg.SetWindowTitle("Loading...")
		msg.SetLabelText("Loading Unicode Tables...")
		msg.SetWindowModality(qt6.WindowModal)
		msg.SetMinimum(0)
		msg.SetMaximum(2)
		msg.SetValue(0)
		msg.Show()
	})

	doneBlocks := make(chan bool)
	go func() {
		b := tables.ParseBlocks()
		blocksMut.Lock()
		blocks = b
		blocksMut.Unlock()
		mainthread.Wait(func() { msg.SetValue(msg.Value() + 1) })

		doneBlocks <- true
	}()

	doneNames := make(chan bool)
	go func() {
		n := tables.ParseNamesList()
		namesMut.Lock()
		names = n
		namesMut.Unlock()
		mainthread.Wait(func() { msg.SetValue(msg.Value() + 1) })

		doneNames <- true
	}()

	<-doneBlocks
	<-doneNames
}

func renderGlyphs() {
	if !tableMut.TryLock() {
		return
	}

	rows := tableWidget.RowCount()
	sheetN := tableScroller.Value() - (rows / 3)
	w := tableWidget.ColumnWidth(0)
	px := max(int(float64(w)*0.6), 4)
	font := qt6.NewQFont2(fontBox.CurrentFont().Family())
	font.SetPixelSize(px)
	rawFont := qt6.QRawFont_FromFont(font)
	normalPalette := tableWidget.Palette()
	inversePalette := qt6.NewQPalette()
	inversePalette.SetColor2(qt6.QPalette__Text, normalPalette.ColorWithCr(qt6.QPalette__Link))
	inversePalette.SetColor2(qt6.QPalette__Base, normalPalette.ColorWithCr(qt6.QPalette__Text))

	for idx := range rows {
		tableWidget.SetRowHeight(idx, w)
		if idx < rows/3 {
			continue
		}

		item := tableWidget.VerticalHeaderItem(idx)
		text := fmt.Sprintf("%03X_", idx+sheetN)
		if item == nil {
			item = qt6.NewQTableWidgetItem2(text)
			tableWidget.SetVerticalHeaderItem(idx, item)
		} else {
			item.SetText(text)
		}

		for col := range 16 {
			char := rune(16*(sheetN+idx) + col)
			cell := tableWidget.CellWidget(idx, col)
			var label *qt6.QLabel
			if cell == nil {
				label = qt6.NewQLabel2()
				label.SetFont(font)
				label.SetAlignment(qt6.AlignCenter)
				cell = label.QWidget
				tableWidget.SetCellWidget(idx, col, label.QWidget)
			} else {
				label = qt6.UnsafeNewQLabel(cell.Metacast("QLabel"))
				label.SetFont(font)
			}

			var targetPalette *qt6.QPalette
			if rawFont.SupportsCharacter(uint(char)) {
				label.SetText(string(char))
				targetPalette = normalPalette
				label.SetPalette(normalPalette)
			} else {
				st := fmt.Sprintf("%04X", int(char))
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
				label.SetText(st)
				label.SetFont(tableWidget.Font())
				targetPalette = inversePalette
			}
			label.SetPalette(targetPalette)
			cell.SetPalette(targetPalette)
		}
	}
	tableMut.Unlock()
}

func main() {
	fmt.Println("hi")
	qt6.NewQApplication(os.Args)
	defer qt6.QApplication_Exec()

	window = qt6.NewQMainWindow(nil)
	window.SetWindowTitle("Glyph Viewer")
	window.SetMinimumSize2(360, 240)

	viewport := qt6.NewQWidget(nil)
	layout := qt6.NewQVBoxLayout(viewport)

	headWidget := qt6.NewQWidget(nil)
	headLayout := qt6.NewQHBoxLayout(headWidget)
	headLayout.SetContentsMargins(0, 0, 0, 0)

	fontBox = qt6.NewQFontComboBox(nil)
	searchBox := qt6.NewQLineEdit(nil)

	headLayout.AddWidget3(searchBox.QWidget, 1, qt6.AlignTop)
	headLayout.AddWidget3(fontBox.QWidget, 0, qt6.AlignTop)

	searchBox.SetPlaceholderText("Search glyphs")

	fontHeight := fontBox.Geometry().Height()
	searchBox.SetFixedHeight(fontHeight)

	layout.AddWidget(headWidget)

	tableWidget = qt6.NewQTableWidget(nil)
	tableWidget.HorizontalHeader().SetSectionResizeMode(qt6.QHeaderView__Fixed)
	tableWidget.VerticalHeader().SetSectionResizeMode(qt6.QHeaderView__Fixed)
	tableWidget.SetHorizontalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	tableWidget.SetVerticalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	tableScroller = qt6.NewQScrollBar2()

	bodyWidget := qt6.NewQWidget(nil)
	bodyLayout := qt6.NewQHBoxLayout(bodyWidget)
	bodyLayout.SetContentsMargins(0, 0, 0, 0)

	bodyLayout.AddWidget(tableWidget.QWidget)
	bodyLayout.AddWidget(tableScroller.QWidget)

	layout.AddWidget(bodyWidget)

	fontBox.OnCurrentFontChanged(func(font *qt6.QFont) {
		rawFont := qt6.QRawFont_FromFont(font)

		fmt.Println("Font changed to", rawFont.FamilyName())
		fmt.Println()
		renderGlyphs()
	})

	window.OnShowEvent(func(_ func(_ *qt6.QShowEvent), evt *qt6.QShowEvent) {
		go func() {
			readTables()
			tableWidget.SetColumnCount(16)
			for c := range 16 {
				item := qt6.NewQTableWidgetItem2(fmt.Sprintf("%X", c))
				tableWidget.SetHorizontalHeaderItem(c, item)
			}
			mainthread.Wait(func() {
				renderGlyphs()
				tableScroller.SetMaximum(int(blocks[len(blocks)-1].End) / 16)
			})
		}()
	})

	tableWidget.OnResizeEvent(func(_ func(_ *qt6.QResizeEvent), evt *qt6.QResizeEvent) {
		tbl_w := contentSize(tableWidget) -
			contentSize(tableWidget.VerticalHeader())

		excess := 16 - (tbl_w % 16)
		col_w := max(tbl_w/16, 4)
		for column := range tableWidget.ColumnCount() {
			if column >= excess {
				tableWidget.SetColumnWidth(column, col_w+1)
			} else {
				tableWidget.SetColumnWidth(column, col_w)
			}
		}
		tableWidget.SetRowCount((tableWidget.Size().Height() / col_w) * 3)
		tableWidget.VerticalScrollBar().SetValue(tableWidget.RowCount() / 3)
		renderGlyphs()
	})

	tableScroller.OnValueChanged(func(value int) {
		renderGlyphs()
	})

	ignoreEvt := false
	tableWidget.OnScrollContentsBy(func(super func(dx int, dy int), dx int, dy int) {
		if ignoreEvt {
			ignoreEvt = false
			return
		}

		ignoreEvt = true
		sheetMax := tableScroller.Maximum()
		sheetMin := tableScroller.Minimum()
		sheetNew := max(min(tableScroller.Value()-dy, sheetMax), sheetMin)
		tableScroller.SetValue(sheetNew)
		top := tableWidget.RowCount() / 3
		tableWidget.VerticalScrollBar().SetValue(top)
	})

	tableWidget.OnKeyPressEvent(func(super func(evt *qt6.QKeyEvent), evt *qt6.QKeyEvent) {
		v := tableScroller.Value()
		top := tableWidget.RowCount() / 3

		dy := top
		if evt.Modifiers()&qt6.ControlModifier > 0 {
			dy = 0x100
		}
		switch evt.Key() {
		case int(qt6.Key_PageDown):
			tableScroller.SetValue(v + dy)
		case int(qt6.Key_PageUp):
			tableScroller.SetValue(v - dy)
		case int(qt6.Key_Down):
			if tableWidget.CurrentRow() >= (top-1)*2 {
				tableScroller.SetValue(v + 1)
				evt.Ignore()
			} else {
				super(evt)
			}
		case int(qt6.Key_Up):
			if tableWidget.CurrentRow() <= top {
				tableScroller.SetValue(v - 1)
				evt.Ignore()
			} else {
				super(evt)
			}
		default:
			super(evt)
		}
	})

	window.SetCentralWidget(viewport)
	window.Show()
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
