package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/db"
)

var ServerList *widget.List
var TabContainer *container.AppTabs
var SelectedServerID int
var Servers []db.Server
var Topics *[]db.Topic
var Subs *[]db.Sub
var CurrentWindow fyne.Window
