package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"fontview/sakura"
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

type FontPair struct {
	Raw  *qt6.QRawFont
	Real *qt6.QFont
}

type Render struct {
	Text  string
	Font  *qt6.QFont
	Style string
}

var (
	fontBox       *qt6.QFontComboBox
	tableWidget   *qt6.QTableWidget
	fontPair      FontPair
	tableScroller *qt6.QScrollBar
	tableMut      sync.Mutex
	window        *qt6.QMainWindow
	msg           *qt6.QProgressDialog
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

var sakurapine sakura.SakuraPalette[string]

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

type FontCache[T any] map[string]map[rune]T

var monoFont *qt6.QFont
var selectedCache = FontCache[Render]{}
var labelCache = FontCache[Render]{}
var fallbackCache = map[rune]string{}
var supportsCache = FontCache[bool]{}

func runeSupported(r rune) bool {
	fam := fontPair.Raw.FamilyName()
	cache := supportsCache[fam]
	if cache == nil {
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
	st := string(r)

	if runeSupported(r) {
		ret.Font = fontPair.Real
		ret.Style = ""
	} else {
		st = runeFallback(r)
		ret.Font = monoFont
		ret.Style = "background-color: " + sakurapine.Hl.Low + ";"
		ret.Style += "color: " + sakurapine.Text.Muted + ";"
	}

	if selected {
		st = fmt.Sprintf("<b><font color='%s'>%s</font></b>", sakurapine.Layer.Base, st)
		ret.Style = ""
	}

	ret.Text = st
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
	}
	tableMut.Unlock()
}

func UpdateRealFont() {
	w := tableWidget.ColumnWidth(0)
	px := max(int(float64(w)*0.6), 4)
	setFont := qt6.NewQFont2(fontBox.CurrentFont().Family())
	setFont.SetPixelSize(px)
	rawFont := qt6.QRawFont_FromFont(setFont)
	fontPair = FontPair{rawFont, setFont}
	for r := range tableWidget.RowCount() {
		for c := range tableWidget.ColumnCount() {
			cell := tableWidget.CellWidget(r, c)
			if cell != nil {
				cell.SetFont(setFont)
			}
		}
	}
}

func main() {
	fmt.Println("hi")
	swatch := sakura.MapSwatch(sakura.Sakura.Parse(), func(c uint) string {
		return fmt.Sprintf("#%06x", c)
	})

	qt6.NewQApplication(os.Args)
	defer qt6.QApplication_Exec()

	monoFont = qt6.QFontDatabase_SystemFont(qt6.QFontDatabase__FixedFont)

	window = qt6.NewQMainWindow(nil)
	window.SetWindowTitle("Glyph Viewer")
	window.SetMinimumSize2(360, 240)

	viewport := qt6.NewQWidget(nil)
	layout := qt6.NewQVBoxLayout(viewport)

	c := viewport.Palette().ColorWithCr(qt6.QPalette__WindowText)
	rgb := sakura.RGB{}.FromHexInt(c.Rgb())
	sum := (rgb.R + rgb.B + rgb.G) / 3
	if sum < 128 {
		sakurapine = swatch.Dawn
	} else {
		sakurapine = swatch.Main
	}

	window.SetStyleSheet(makeStyleSheet(map[string]map[string]string{
		"QMainWindow": {
			"background-color": sakurapine.Layer.Base,
			"color":            sakurapine.Text.Normal,
		},
		"QTableWidget": {
			"selection-color":            sakurapine.Layer.Base,
			"selection-background-color": sakurapine.Paint.Rose,
		},
	}))

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
	tableWidget.SetVerticalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	tableScroller = qt6.NewQScrollBar2()

	tableWidget.HorizontalHeader().SetFont(monoFont)

	bodyWidget := qt6.NewQWidget(nil)
	bodyLayout := qt6.NewQHBoxLayout(bodyWidget)
	bodyLayout.SetContentsMargins(0, 0, 0, 0)

	bodyLayout.AddWidget(tableWidget.QWidget)
	bodyLayout.AddWidget(tableScroller.QWidget)

	layout.AddWidget(bodyWidget)

	fontBox.OnCurrentFontChanged(func(font *qt6.QFont) {
		UpdateRealFont()
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
				tableScroller.SetValue(0)
			})
		}()
	})

	autoSize := true
	col_w := 0

	resizeGlyphs := func() {
		excess := 16
		if autoSize {
			tbl_w := contentSize(tableWidget) - 8 -
				contentSize(tableWidget.VerticalHeader())
			col_w = max(tbl_w/16, 4)
			excess -= (tbl_w % 16)
		}

		for column := range tableWidget.ColumnCount() {
			if column >= excess {
				tableWidget.SetColumnWidth(column, col_w+1)
			} else {
				tableWidget.SetColumnWidth(column, col_w)
			}
		}
		tableWidget.SetRowCount((tableWidget.Size().Height() / col_w) * 3)
		tableWidget.VerticalScrollBar().SetValue(tableWidget.RowCount() / 3)
		UpdateRealFont()
		renderGlyphs()
	}

	tableWidget.OnResizeEvent(func(_ func(_ *qt6.QResizeEvent), evt *qt6.QResizeEvent) {
		resizeGlyphs()
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
		super(dx, 0)
	})

	tableWidget.OnKeyPressEvent(func(super func(evt *qt6.QKeyEvent), evt *qt6.QKeyEvent) {
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
				autoSize = false
				col_w += 1
				resizeGlyphs()
			}
		case qt6.Key_Minus:
			if mods&qt6.ControlModifier > 0 {
				autoSize = false
				col_w = max(col_w-1, 4)
				resizeGlyphs()
			}
		case qt6.Key_0:
			if mods&qt6.ControlModifier > 0 {
				autoSize = true
				resizeGlyphs()
			}

		default:
			super(evt)
		}
	})

	tableWidget.OnCurrentCellChanged(func(_, _, _, _ int) {
		renderGlyphs()
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
