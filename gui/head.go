package gui

import "github.com/mappu/miqt/qt6"

var (
	btnBack *qt6.QPushButton
	btnFwd  *qt6.QPushButton
)

func MakeHead() *qt6.QWidget {
	headWidget := qt6.NewQWidget(nil)
	headLayout := qt6.NewQHBoxLayout(headWidget)
	headLayout.SetContentsMargins(0, 0, 0, 0)

	fontBox = qt6.NewQFontComboBox(nil)
	searchBox := qt6.NewQLineEdit(nil)

	btnBack = qt6.NewQPushButton2()
	btnBack.SetIcon(icons["go-previous"])
	btnBack.SetDisabled(true)
	btnBack.OnClicked(btn_BackEvt)

	btnFwd = qt6.NewQPushButton2()
	btnFwd.SetIcon(icons["go-next"])
	btnFwd.SetDisabled(true)
	btnFwd.OnClicked(btn_FwdEvt)

	headLayout.AddWidget3(btnBack.QWidget, 0, qt6.AlignTop)
	headLayout.AddWidget3(btnFwd.QWidget, 0, qt6.AlignTop)
	headLayout.AddWidget3(searchBox.QWidget, 1, qt6.AlignTop)
	headLayout.AddWidget3(fontBox.QWidget, 0, qt6.AlignTop)

	searchBox.SetPlaceholderText("Search glyphs")

	fontHeight := fontBox.Geometry().Height()
	searchBox.SetFixedHeight(fontHeight)

	fontBox.OnCurrentFontChanged(func(font *qt6.QFont) {
		UpdateRealFont()
	})

	return headWidget
}
