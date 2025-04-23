package main

import (
	"fmt"
	"io"
	"os"

	"fontview/tables"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"golang.org/x/image/font/sfnt"
)

func readFont(path string) (*sfnt.Font, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return sfnt.Parse(data)
}

func main() {
	qt6.NewQApplication(os.Args)
	defer qt6.QApplication_Exec()

	msg := qt6.NewQProgressDialog(nil)
	msg.SetWindowTitle("Loading...")
	msg.SetLabelText("Loading Unicode Tables...")
	msg.SetWindowModality(qt6.WindowModal)
	msg.SetMinimum(0)
	msg.SetMaximum(2)
	msg.SetValue(0)
	msg.Show()

	go func() {
		tables.ParseBlocks()
		mainthread.Wait(func() { msg.SetValue(1) })

		tables.ParseNamesList()
		mainthread.Wait(func() { msg.SetValue(2) })
	}()

	window := qt6.NewQMainWindow(nil)
	window.SetWindowTitle("Glyph Viewer")
	window.SetMinimumSize2(360, 240)

	viewport := qt6.NewQWidget(nil)
	layout := qt6.NewQVBoxLayout(viewport)

	headWidget := qt6.NewQWidget(nil)
	headLayout := qt6.NewQHBoxLayout(headWidget)

	fontBox := qt6.NewQFontComboBox(nil)
	searchBox := qt6.NewQLineEdit(nil)

	headLayout.AddWidget3(searchBox.QWidget, 1, qt6.AlignTop)
	headLayout.AddWidget3(fontBox.QWidget, 0, qt6.AlignTop)

	searchBox.SetPlaceholderText("Search glyphs")

	fontHeight := fontBox.Geometry().Height()
	searchBox.SetFixedHeight(fontHeight)

	layout.AddWidget3(headWidget, 0, qt6.AlignTop)

	fontBox.OnCurrentFontChanged(func(font *qt6.QFont) {
		fmt.Println("Font changed to", font.Style)
		fmt.Println()
	})

	window.SetCentralWidget(viewport)
	window.Show()
}
