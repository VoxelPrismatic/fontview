package main

import (
	"fmt"
	"os"
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
		fmt.Println("Blocked")
		return
	} else {
		fmt.Println("Lock acquired")
	}

	rows := tableWidget.RowCount()
	sheetN := tableScroller.Value() - 1
	w := tableWidget.ColumnWidth(0)
	px := max(int(float64(w)*0.6), 4)
	font := qt6.NewQFont2(fontBox.CurrentFont().Family())
	font.SetPixelSize(px)
	rawFont := qt6.QRawFont_FromFont(font)

	for idx := range rows {
		tableWidget.SetRowHeight(idx, w)
		if idx == 0 {
			continue
		}

		item := tableWidget.VerticalHeaderItem(idx)
		text := fmt.Sprintf("u%03x_", idx+sheetN)
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

			if rawFont.SupportsCharacter(uint(char)) {
				label.SetText(string(char))
				label.SetForegroundRole(qt6.QPalette__Text)
				label.SetBackgroundRole(qt6.QPalette__Text)
			} else {
				label.SetText(fmt.Sprintf("u%04x", int(char)))
				label.SetFont(tableWidget.Font())
				label.SetForegroundRole(qt6.QPalette__Dark)
				label.SetBackgroundRole(qt6.QPalette__Dark)
				cell.SetBackgroundRole(qt6.QPalette__AlternateBase)
				cell.SetForegroundRole(qt6.QPalette__LinkVisited)
				label.SetFixedSize2(w, w)
			}
		}
	}
	fmt.Println("Unlocked")
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
	tableWidget.SetHorizontalHeaderLabels([]string{
		"0", "1", "2", "3", "4", "5", "6", "7",
		"8", "9", "a", "b", "c", "d", "e", "f",
	})
	tableWidget.HorizontalHeader().SetSectionResizeMode(qt6.QHeaderView__Fixed)
	tableWidget.VerticalHeader().SetSectionResizeMode(qt6.QHeaderView__Fixed)
	tableWidget.SetHorizontalScrollBarPolicy(qt6.ScrollBarAlwaysOff)

	// tableWidget.SetVerticalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
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
			tableScroller.SetMaximum(int(blocks[len(blocks)-1].End) / 16)
			tableWidget.SetColumnCount(16)
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
		tableWidget.SetRowCount(tableWidget.Size().Height()/col_w + 4)
		tableWidget.VerticalScrollBar().SetValue(1)
		renderGlyphs()
	})

	tableScroller.OnValueChanged(func(value int) {
		fmt.Printf("Scroll to : %d\n", value)
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
		fmt.Printf("(%d, %d)\n", dx, dy)
		sheetNew := max(min(tableScroller.Value()-dy, sheetMax), sheetMin)
		tableScroller.SetValue(sheetNew)
		tableWidget.VerticalScrollBar().SetValue(1)
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
