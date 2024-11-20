package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/db"
	"github.com/devalexandre/brokers-ui/messaging"
)

var BrokerServers = make(map[int]messaging.MessagingClient)
var BrokerError error

func AddServer(window fyne.Window, db *db.Database) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter server name...")
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter server URL...")

	clientTypeEntry := widget.NewEntry()
	clientTypeEntry.SetPlaceHolder("Enter client type...")
	clientTypeEntry.Disable()

	selectClientType := widget.NewSelect(messaging.ClientType, func(selected string) {
		clientTypeEntry.SetText(selected)
	})

	dialog := dialog.NewForm(
		"Add Server",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Server Name", nameEntry),
			widget.NewFormItem("Server URL", urlEntry),
			widget.NewFormItem("Client Type", selectClientType),
		},
		func(confirmed bool) {
			if confirmed {
				db.SaveServer(nameEntry.Text, urlEntry.Text, clientTypeEntry.Text)
				db.LoadServers()           // Recarrega a lista de servidores
				Servers = db.LoadServers() // Recarrega a lista de servidores
				ServerList.Refresh()       // Atualiza a lista exibida
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func DisplayServerOptions(window fyne.Window, database *db.Database, server db.Server) {
	ClearTabs()

	menu := container.NewHBox(
		widget.NewButtonWithIcon("Add Topic", theme.ContentAddIcon(), func() {
			AddTopic(window, database, server.ID)
		}),
		widget.NewButtonWithIcon("Add Sub", theme.ContentAddIcon(), func() {
			AddSub(window, database, server.ID)
		}),
		widget.NewButtonWithIcon("Disconnect", theme.MediaStopIcon(), func() {
			DisconnectFromServer()
		}),
	)

	editButton := widget.NewButtonWithIcon("Edit Connection", theme.ViewRefreshIcon(), func() {
		EditServerConnection(window, database, server)
	})

	panel := container.NewVBox(
		menu,
		widget.NewLabel(fmt.Sprintf("Connected to %s (%s)", server.Name, server.URL)),
		editButton,
	)

	log.Printf("Connecting to NATS server: %s", server.URL)
	log.Printf("Selected server %v", server.ID)

	switch server.Client {
	case "NATS":
		BrokerServers[server.ID], BrokerError = messaging.NewNats(server.URL)
	}
	if BrokerError != nil {
		log.Printf("Error connecting to NATS server: %v", BrokerError)
		dialog.ShowError(BrokerError, window)
	} else {
		log.Printf("Connected to NATS server")
		log.Printf("Selected server %v", server.ID)
		RefreshTopicsAndSubs(server.ID, database)
	}

	configTab := container.NewTabItem("Config", panel)
	TabContainer.Append(configTab)
	TabContainer.Select(configTab)
	TabContainer.Refresh()

	AddDashboardTab(database)
	AddTabsForTopicsAndSubs(window, database, server.ID)
}

func ClearTabs() {
	TabContainer.Items = []*container.TabItem{}
	TabContainer.Refresh()
}

func DisconnectFromServer() {
	if client, ok := BrokerServers[SelectedServerID]; ok {
		client.Close()
		delete(BrokerServers, SelectedServerID)
	}
	ClearTabs()
	TabContainer.Append(container.NewTabItem("Welcome", widget.NewLabel("Welcome to the NATS Client!")))
	TabContainer.Refresh()
}

func EditServerConnection(window fyne.Window, db *db.Database, server db.Server) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(server.Name)
	urlEntry := widget.NewEntry()
	urlEntry.SetText(server.URL)

	dialog := dialog.NewForm(
		"Edit Server Connection",
		"Save",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Server Name", nameEntry),
			widget.NewFormItem("Server URL", urlEntry),
		},
		func(confirmed bool) {
			if confirmed {
				db.UpdateServer(server.ID, nameEntry.Text, urlEntry.Text)
				db.LoadServers()
				ServerList.Refresh()
				DisplayServerOptions(window, db, server)
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func RefreshTopicsAndSubs(serverID int, db *db.Database) {
	topics := db.LoadTopics(serverID)
	subs := db.LoadSubs(serverID)

	Topics = &topics
	Subs = &subs
}
