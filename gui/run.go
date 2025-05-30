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
	fontBox    *qt6.QFontComboBox
	fontPair   FontPair
	window     *qt6.QMainWindow
	sakurapine sakura.SakuraPalette[string]
	monoFont   *qt6.QFont
	icons      map[string]*qt6.QIcon
)

func Launch() {
	qt6.NewQApplication(os.Args)
	defer qt6.QApplication_Exec()

	monoFont = qt6.QFontDatabase_SystemFont(qt6.QFontDatabase__FixedFont)

	window = qt6.NewQMainWindow(nil)
	window.SetWindowTitle("Glyph Viewer")
	window.SetMinimumSize2(360, 240)

	determineTheme()

	viewport := qt6.NewQWidget(nil)
	layout := qt6.NewQVBoxLayout(viewport)

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

	layout.AddWidget(MakeHead())
	layout.AddWidget(MakeTable())
	window.AddDockWidget(qt6.RightDockWidgetArea, MakeInfo())

	window.OnShowEvent(func(_ func(_ *qt6.QShowEvent), evt *qt6.QShowEvent) {
		go boot()
	})

	window.SetCentralWidget(viewport)
	window.Show()
}

var searchCache = map[string]map[string][]rune{}

func boot() {
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
		onLink("0")
	})
}

func determineTheme() {
	swatch := sakura.MapSwatch(sakura.Sakura.Parse(), func(c uint) string {
		return fmt.Sprintf("#%06x", c)
	})

	c := window.Palette().ColorWithCr(qt6.QPalette__WindowText)
	rgb := sakura.RGB{}.FromHexInt(c.Rgb())
	sum := (rgb.R + rgb.B + rgb.G) / 3
	if sum < 128 {
		sakurapine = swatch.Dawn
	} else {
		sakurapine = swatch.Main
	}

	icons = map[string]*qt6.QIcon{}

	for name, svg := range kdeIcons {
		svg = fmt.Sprintf(svg, sakurapine.Text.Normal)
		img := qt6.QImage_FromData5([]byte(svg), "image/svg+xml")
		pix := qt6.QPixmap_FromImage(img)
		ico := qt6.NewQIcon2(pix)
		icons[name] = ico
	}

}
