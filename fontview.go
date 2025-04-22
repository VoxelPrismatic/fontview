package main

import (
	"os"

	"github.com/mappu/miqt/qt6"
)

func main() {
	qt6.NewQApplication(os.Args)
	defer qt6.QApplication_Exec()

	window := qt6.NewQMainWindow(nil)
	window.SetWindowTitle("Glyph Viewer")
	window.SetMinimumSize2(360, 240)

	viewport := qt6.NewQWidget(nil)
	layout := qt6.NewQVBoxLayout(viewport)

	headWidget := qt6.NewQWidget(nil)
	headLayout := qt6.NewQHBoxLayout(headWidget)

	fontBox := qt6.NewQFontComboBox(nil)
	searchBox := qt6.NewQLineEdit(nil)
	sizeBox := qt6.NewQSpinBox(nil)

	headLayout.AddWidget3(searchBox.QWidget, 1, qt6.AlignTop)
	headLayout.AddWidget3(fontBox.QWidget, 0, qt6.AlignTop)
	headLayout.AddWidget3(sizeBox.QWidget, 0, qt6.AlignTop)

	searchBox.SetPlaceholderText("Search glyphs")
	sizeBox.SetMinimum(8)
	sizeBox.SetMaximum(72)
	sizeBox.SetValue(24)

	fontHeight := fontBox.Geometry().Height()
	sizeBox.SetFixedHeight(fontHeight)
	searchBox.SetFixedHeight(fontHeight)

	filterWidget := qt6.NewQWidget(nil)
	filterLayout := qt6.NewQHBoxLayout(filterWidget)
	filterLayout.AddStretch()

	var monospaceFilter qt6.QFontComboBox__FontFilter
	var bitmapFilter qt6.QFontComboBox__FontFilter

	filterByMonospace := qt6.NewQCheckBox(nil)
	filterByMonospace.SetText("Monospace")
	filterByMonospace.SetTristate()
	filterByMonospace.OnStateChanged(func(state int) {
		switch qt6.CheckState(state) {
		case qt6.Checked:
			filterByMonospace.SetToolTip("Show only monospace fonts")
			monospaceFilter = qt6.QFontComboBox__MonospacedFonts
		case qt6.PartiallyChecked:
			filterByMonospace.SetToolTip("Show all fonts")
			monospaceFilter = qt6.QFontComboBox__AllFonts
		case qt6.Unchecked:
			filterByMonospace.SetToolTip("Show only proportional fonts")
			monospaceFilter = qt6.QFontComboBox__ProportionalFonts
		}
		fontBox.SetFontFilters(monospaceFilter | bitmapFilter)
	})
	filterByMonospace.SetCheckState(qt6.PartiallyChecked)
	filterLayout.AddWidget3(filterByMonospace.QWidget, 0, qt6.AlignTop)

	filterByBitmap := qt6.NewQCheckBox(nil)
	filterByBitmap.SetText("Bitmap")
	filterByBitmap.SetTristate()
	filterByBitmap.OnStateChanged(func(state int) {
		switch qt6.CheckState(state) {
		case qt6.Checked:
			filterByBitmap.SetToolTip("Show only bitmap fonts")
			bitmapFilter = qt6.QFontComboBox__NonScalableFonts
		case qt6.PartiallyChecked:
			filterByBitmap.SetToolTip("Show all fonts")
			bitmapFilter = qt6.QFontComboBox__AllFonts
		case qt6.Unchecked:
			filterByBitmap.SetToolTip("Show only vector fonts")
			bitmapFilter = qt6.QFontComboBox__ScalableFonts
		}
		fontBox.SetFontFilters(monospaceFilter | bitmapFilter)
	})
	filterByBitmap.SetCheckState(qt6.PartiallyChecked)
	filterLayout.AddWidget3(filterByBitmap.QWidget, 0, qt6.AlignTop)

	layout.AddWidget3(headWidget, 0, qt6.AlignTop)
	layout.AddWidget3(filterWidget, 0, qt6.AlignTop)

	window.SetCentralWidget(viewport)
	window.Show()
}
