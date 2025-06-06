package gui

import (
	"fmt"
	"fontview/tables"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type QPtr interface {
	Metacast(string) unsafe.Pointer
}

type GroupBox[T QPtr] struct {
	widget T
	group  *qt6.QGroupBox
	layout *qt6.QVBoxLayout
}

func (g *GroupBox[T]) Init(title string, widget T) T {
	g.group = qt6.NewQGroupBox3(title)
	g.layout = qt6.NewQVBoxLayout(g.group.QWidget)
	g.widget = widget
	w := qt6.UnsafeNewQWidget(widget.Metacast("QWidget"))
	g.layout.AddWidget(w)
	return widget
}

var (
	infoWidget *qt6.QWidget
	infoPanel  *qt6.QDockWidget
	infoLayout *qt6.QBoxLayout

	info_Tree     GroupBox[*qt6.QTreeWidget]
	info_Tab      *qt6.QTabWidget
	tree_AltNames *qt6.QTreeWidgetItem
	tree_Remarks  *qt6.QTreeWidgetItem
	tree_Refs     *qt6.QTreeWidgetItem
	tree_Approx   *qt6.QTreeWidgetItem
	tree_Equiv    *qt6.QTreeWidgetItem

	info_Preview  GroupBox[*qt6.QLabel]
	info_Details  GroupBox[*qt6.QWidget]
	info_RawBlock GroupBox[*qt6.QLabel]

	info_CodeSelector  *qt6.QComboBox
	info_CodeLabel     *qt6.QLabel
	info_BlockLabel    *qt6.QLabel
	info_NameLabel     *qt6.QLabel
	info_CategoryLabel *qt6.QLabel

	curNode tables.Node
	caser   = cases.Title(language.English)
)

func MakeInfo() *qt6.QDockWidget {
	infoWidget = qt6.NewQWidget2()
	infoLayout = qt6.NewQBoxLayout2(qt6.QBoxLayout__Down, infoWidget)
	infoPanel = qt6.NewQDockWidget2("Glyph")

	infoPanel.SetAllowedAreas(
		qt6.BottomDockWidgetArea |
			qt6.RightDockWidgetArea |
			qt6.LeftDockWidgetArea,
	)

	infoLayout.AddWidget(makeInfo_Preview())
	infoLayout.AddWidget(makeInfo_Details())
	infoLayout.AddWidget2(makeInfo_Tab(), 1)

	timer := qt6.NewQTimer()
	timer.OnTimerEvent(func(super func(evt *qt6.QTimerEvent), evt *qt6.QTimerEvent) {
		mainthread.Start(updateInfo)
	})
	timer.SetInterval(100)
	timer.Start(100)

	scrollArea := qt6.NewQScrollArea2()
	scrollArea.SetWidgetResizable(true)
	scrollArea.SetWidget(infoWidget)
	infoPanel.SetWidget(scrollArea.QWidget)

	return infoPanel
}

func updateInfo() {
	rows := tableWidget.RowCount()
	curCell := tableWidget.CurrentIndex()
	if curCell == nil {
		return
	}

	curR, curC := curCell.Row(), curCell.Column()
	sheetN := tableScroller.Value() - (rows / 3)
	char := rune(16*(sheetN+curR) + curC)
	node := names[fmt.Sprintf("%04X", char)]
	if node == nil {
		if names == nil || blocks == nil {
			return
		}
		node = &tables.Node{
			Point: char,
			Name:  "Undefined",
			Code:  fmt.Sprintf("%04X", char),
			Remarks: []string{
				"This character is not defined by the unicode spec",
			},
			Block: blocks[len(blocks)-1],
			Raw:   "<Undefined>",
		}
		namesMut.Lock()
		names[node.Code] = node
		namesMut.Unlock()
		blocksMut.Lock()
		for _, block := range blocks {
			if block.Start <= char && char <= block.End {
				block.Nodes = append(block.Nodes, node)
				if (node.Block.End - node.Block.Start) > (block.End - block.Start) {
					node.Block = block
				}
			}
		}
		blocksMut.Unlock()

		return
	}
	if node.Code == curNode.Code {
		return
	}
	curNode = *node

	updateInfo_Preview(*node)
	updateInfo_List(*node)
	updateInfo_Details(*node)
	updateInfo_RawBlock(*node)
}

func make_Label(label string) *qt6.QLabel {
	ret := qt6.NewQLabel3(label)
	ret.SetWordWrap(true)
	ret.SetAlignment(qt6.AlignVCenter | qt6.AlignRight)
	ret.SetTextInteractionFlags(qt6.TextSelectableByMouse)
	return ret
}

func makeInfo_Tab() *qt6.QWidget {
	info_Tab = qt6.NewQTabWidget2()
	info_Tab.AddTab(makeInfo_List(), "Metadata")
	info_Tab.AddTab(makeInfo_RawBlock(), "Raw Data")
	info_Tab.SetTabPosition(qt6.QTabWidget__South)
	info_Tab.SetDocumentMode(true)
	return info_Tab.QWidget
}

func makeInfo_List() *qt6.QWidget {
	tree := info_Tree.Init("Unicode Metadata", qt6.NewQTreeWidget2())
	tree_AltNames = qt6.NewQTreeWidgetItem()
	tree_AltNames.SetText(0, "Alternate Names")
	tree_Approx = qt6.NewQTreeWidgetItem()
	tree_Approx.SetText(0, "Alternates")
	tree_Equiv = qt6.NewQTreeWidgetItem()
	tree_Equiv.SetText(0, "Equivalents")
	tree_Refs = qt6.NewQTreeWidgetItem()
	tree_Refs.SetText(0, "References")
	tree_Remarks = qt6.NewQTreeWidgetItem()
	tree_Remarks.SetText(0, "Remarks")
	tree.SetColumnCount(2)
	tree.SetHeaderHidden(true)

	tree.AddTopLevelItems([]*qt6.QTreeWidgetItem{
		tree_AltNames,
		tree_Remarks,
		tree_Refs,
		tree_Equiv,
		tree_Approx,
	})

	tree.OnResizeEvent(func(super func(event *qt6.QResizeEvent), event *qt6.QResizeEvent) {
		tree.SetColumnWidth(0, event.Size().Width()-32)
		tree.SetColumnWidth(1, 32)
	})
	return info_Tree.group.QWidget
}

func updateInfo_List(node tables.Node) {
	updateList_Generic(node.AltNames, tree_AltNames)
	updateList_Generic(node.Approx, tree_Approx)
	updateList_Generic(node.Equiv, tree_Equiv)
	updateList_Generic(node.Refs, tree_Refs)
	updateList_Generic(node.Remarks, tree_Remarks)
}

func updateList_Generic(stuff []string, target *qt6.QTreeWidgetItem) {
	for range target.ChildCount() {
		target.RemoveChild(target.Child(0))
	}

	var color *qt6.QColor
	if len(stuff) == 0 {
		color = qt6.NewQColor6(sakurapine.Text.Muted)
	} else {
		color = qt6.NewQColor6(sakurapine.Text.Normal)
	}

	target.SetForeground(0, qt6.NewQBrush3(color))
	target.SetForeground(1, qt6.NewQBrush3(color))
	target.SetText(1, fmt.Sprintf("%d", len(stuff)))

	for _, item := range stuff {
		child := qt6.NewQTreeWidgetItem()
		child.SetFlags(qt6.ItemNeverHasChildren)
		label := qt6.NewQLabel3(render_List(item))
		label.SetWordWrap(true)
		label.OnLinkActivated(onLink)
		label.SetTextFormat(qt6.RichText)
		label.SetStyleSheet("padding-left: 2px;")
		target.AddChild(child)
		info_Tree.widget.SetItemWidget(child, 0, label.QWidget)
	}

	target.SetExpanded(true)
}
func render_List(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = link.ReplaceAllStringFunc(s, func(u string) string {
		node, ok := names[u]
		var title string
		if !ok || node == nil {
			title = "&lt;Undefined&gt;"
		} else {
			title = caser.String(names[u].Name)
		}
		i, err := strconv.ParseInt(u, 16, 64)
		if err != nil {
			panic(err)
		}

		return fmt.Sprintf("<a href=\"%d\" style=\"margin:0px;padding:0px;\">%s: %s</a>", i, u, title)
	})
	return s
}

func makeInfo_Details() *qt6.QWidget {
	gridWidget := qt6.NewQWidget2()
	grid := qt6.NewQGridLayout(gridWidget)
	grid.SetContentsMargins(0, 0, 0, 0)
	info_Details.Init("Details", gridWidget)

	item := qt6.NewQLabel3("<b>Name</b>")
	info_NameLabel = make_Label("Name")
	grid.AddWidget4(item.QWidget, 0, 0, qt6.AlignLeft)
	grid.AddWidget4(info_NameLabel.QWidget, 0, 1, qt6.AlignRight)

	item = qt6.NewQLabel3("<b>Block</b>")
	info_BlockLabel = make_Label("Block")
	grid.AddWidget4(item.QWidget, 1, 0, qt6.AlignLeft)
	grid.AddWidget4(info_BlockLabel.QWidget, 1, 1, qt6.AlignRight)

	item = qt6.NewQLabel3("<b>Category</b>")
	info_CategoryLabel = make_Label("Category")
	grid.AddWidget4(item.QWidget, 2, 0, qt6.AlignLeft)
	grid.AddWidget4(info_CategoryLabel.QWidget, 2, 1, qt6.AlignRight)

	info_CodeWidget := qt6.NewQWidget2()
	info_CodeLayout := qt6.NewQHBoxLayout(info_CodeWidget)
	info_CodeLayout.SetContentsMargins(0, 0, 0, 0)

	info_CodeSelector = qt6.NewQComboBox2()
	info_CodeCopy := qt6.NewQPushButton2()
	if copyErr != nil {
		info_CodeCopy.SetIcon(icons["apport"])
		info_CodeCopy.SetToolTip(copyErr.Error())
		info_CodeCopy.SetDisabled(true)
	} else {
		info_CodeCopy.SetIcon(icons["edit-copy"])
		info_CodeCopy.SetToolTip(copyCmd)
	}
	info_CodeLabel = make_Label("Code point")

	info_CodeLayout.AddWidget(info_CodeSelector.QWidget)
	info_CodeLayout.AddWidget(info_CodeCopy.QWidget)
	grid.AddWidget4(info_CodeWidget, 3, 0, qt6.AlignLeft)
	grid.AddWidget4(info_CodeLabel.QWidget, 3, 1, qt6.AlignRight)

	keys := []string{}
	for key := range tables.CodeEncoder {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	info_CodeSelector.AddItems(keys)

	info_CodeSelector.OnCurrentTextChanged(func(selected string) {
		formatted := tables.CodeEncoder[selected](curNode.Point)
		info_CodeLabel.SetText(formatted)
	})
	gridWidget.OnResizeEvent(func(super func(event *qt6.QResizeEvent), event *qt6.QResizeEvent) {
		w := event.Size().Width() -
			info_CodeSelector.Size().Width() -
			info_CodeCopy.Size().Width() - 32
		info_NameLabel.SetMinimumWidth(w)
		info_BlockLabel.SetMinimumWidth(w)
		info_CategoryLabel.SetMinimumWidth(w)
		info_CodeLabel.SetMinimumWidth(w)
	})

	checkTimer := qt6.NewQTimer()
	checkTimer.OnTimerEvent(func(super func(param1 *qt6.QTimerEvent), param1 *qt6.QTimerEvent) {
		checkTimer.Stop()
		info_CodeCopy.SetIcon(icons["edit-copy"])
	})
	info_CodeCopy.OnClicked(func() {
		info_CodeCopy.SetIcon(icons["view-refresh"])
		err := copyRune()
		if err != nil {
			info_CodeCopy.SetIcon(icons["apport"])
			info_CodeCopy.SetToolTip(err.Error())
		} else {
			info_CodeCopy.SetIcon(icons["checkbox"])
			checkTimer.Start(1000)
		}
	})

	return info_Details.group.QWidget
}

func updateInfo_Details(node tables.Node) {
	info_NameLabel.SetText(caser.String(node.Name))
	info_BlockLabel.SetText(node.Block.Name)
	classes := map[string][]string{}
	for cat, block := range unicode.Categories {
		if unicode.In(node.Point, block) {
			class, ok := tables.CategoryMap[string(cat[0])]
			if !ok {
				if classes["Unclassified"] == nil {
					classes["Unclassified"] = []string{cat}
				} else {
					classes["Unclassified"] = append(classes["Unclassified"], cat)
				}
				continue
			}

			if _, ok := classes[class["!"]]; !ok {
				classes[class["!"]] = []string{}
			}

			if len(cat) == 1 {
				continue
			}

			name, ok := class[string(cat[1])]
			if !ok {
				name = fmt.Sprintf("<%s>", cat)
			}
			classes[class["!"]] = append(classes[class["!"]], name)
		}
	}

	categories := []string{}
	for key, val := range classes {
		if len(val) == 0 {
			categories = append(categories, key)
		} else {
			category := fmt.Sprintf("%s: %s", key, strings.Join(val, ", "))
			categories = append(categories, category)
		}
	}
	info_CategoryLabel.SetText(strings.Join(categories, "; "))
	selected := info_CodeSelector.CurrentText()
	formatted := tables.CodeEncoder[selected](node.Point)
	info_CodeLabel.SetText(formatted)
}

func makeInfo_Preview() *qt6.QWidget {
	label := info_Preview.Init("Preview", qt6.NewQLabel2())
	label.SetAlignment(qt6.AlignCenter)
	label.SetWordWrap(true)
	label.SetTextInteractionFlags(qt6.TextSelectableByMouse)
	label.OnMousePressEvent(func(super func(ev *qt6.QMouseEvent), ev *qt6.QMouseEvent) {
		copySym()
	})
	return info_Preview.group.QWidget
}

func updateInfo_Preview(node tables.Node) {
	font := qt6.NewQFont5(fontPair.Real)
	font.SetPixelSize(156)
	font.SetStyleStrategy(qt6.QFont__NoFontMerging)

	info_Preview.widget.SetFont(font)
	info_Preview.widget.SetText(string(node.Point))
}

func makeInfo_RawBlock() *qt6.QWidget {
	label := info_RawBlock.Init("Raw Unicode Data", qt6.NewQLabel2())
	label.SetAlignment(qt6.AlignTop | qt6.AlignLeft)
	label.SetFont(monoFont)
	label.SetWordWrap(true)
	label.SetTextInteractionFlags(qt6.TextSelectableByMouse | qt6.LinksAccessibleByMouse)
	label.OnLinkActivated(onLink)
	return info_RawBlock.group.QWidget
}

func updateInfo_RawBlock(node tables.Node) {
	s := strings.ReplaceAll(node.Raw, "<", "&lt;")
	s = link.ReplaceAllStringFunc(s, func(u string) string {
		i, err := strconv.ParseInt(u, 16, 64)
		if err != nil {
			panic(err)
		}

		return fmt.Sprintf("<a href=\"%d\">%s</a>", i, u)
	})
	s = strings.ReplaceAll(s, "\n\t", "<br>"+strings.Repeat("&nbsp;", len(node.Code)+1))
	s = strings.ReplaceAll(s, "\n", "<br>")
	info_RawBlock.widget.SetText(s)
}

var link = regexp.MustCompile(`\b([A-F0-9]{4,})\b`)
