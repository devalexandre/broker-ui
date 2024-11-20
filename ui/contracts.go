package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/db"
)

var ServerList *widget.List
var TopicsDropdown *widget.Select
var SubsDropdown *widget.Select
var TabContainer *container.AppTabs
var SelectedServerID int
var Servers []db.Server
