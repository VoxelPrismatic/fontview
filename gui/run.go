package gui

import (
	"fmt"
	"os"

	"fontview/sakura"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
)

type FontPair struct {
	Raw  *qt6.QRawFont
	Real *qt6.QFont
}

var (
	fontBox  *qt6.QFontComboBox
	fontPair FontPair
	window   *qt6.QMainWindow
	msg      *qt6.QProgressDialog
)

var sakurapine sakura.SakuraPalette[string]

var monoFont *qt6.QFont

func Launch() {
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

	tableWidget.OnResizeEvent(func(_ func(_ *qt6.QResizeEvent), evt *qt6.QResizeEvent) {
		labelCache = FontCache[Render]{}
		selectedCache = FontCache[Render]{}
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

	tableWidget.OnKeyPressEvent(tbl_OnKeyEvt)

	tableWidget.OnCurrentCellChanged(func(_, _, _, _ int) {
		renderGlyphs()
	})

	window.SetCentralWidget(viewport)
	window.Show()
}

var searchCache = map[string]map[string][]rune{}
