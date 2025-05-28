package gui

import (
	"fontview/tables"
	"sync"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
)

var (
	blocks    []tables.Block
	blocksMut sync.Mutex
	names     map[string]tables.Node
	namesMut  sync.Mutex
	msg       *qt6.QProgressDialog
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
