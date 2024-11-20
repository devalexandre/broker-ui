package main

import (
	"fmt"
	"fyne.io/fyne/v2/theme"
	"log"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/db"
	"github.com/devalexandre/brokers-ui/themes"
	"github.com/devalexandre/brokers-ui/ui"
)

func main() {
	// Initialize the application
	myApp := app.New()
	myWindow := myApp.NewWindow("Broker Client")

	database, err := db.NewDatabase()
	if err != nil {
		panic(err)
	}

	// Retrieve dark mode setting
	darkModeValue, err := database.GetSetting("dark_mode")
	if err != nil {
		log.Printf("Error retrieving dark mode setting: %v", err)
	}

	isDarkMode := false
	if darkModeValue == "true" {
		isDarkMode = true
		myApp.Settings().SetTheme(themes.DraculaTheme{})
	} else {
		myApp.Settings().SetTheme(themes.LightTheme{})
	}

	// Initialize TabContainer
	ui.TabContainer = container.NewAppTabs()

	// Welcome message
	welcomeMessage := widget.NewRichTextFromMarkdown(`# Welcome to Broker Client

This application allows you to connect to **NATS, Kafka** servers, create topics, and subscribe to subjects.

- **Add Server**: Add a new NATS/Kafka server connection.
- **Topics**: Publish messages to topics.
- **Subscriptions**: Receive messages from subjects.

Developed by [Alexandre E Souza](https://www.linkedin.com/in/devevantelista)`)
	// Add welcome tab
	welcomeTab := container.NewTabItem("Welcome", welcomeMessage)
	ui.TabContainer.Append(welcomeTab)

	// Load servers from the database
	ui.Servers = database.LoadServers()

	slog.Debug("Servers: %v", ui.Servers)

	// Initialize Server List with delete icon
	ui.ServerList = widget.NewList(
		func() int { return len(ui.Servers) },
		func() fyne.CanvasObject {
			// Create a label and a delete button for each server item
			label := widget.NewLabel("Server Name")
			deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			return container.NewHBox(label, deleteButton)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			// Update the label text and set the delete action
			itemContainer := o.(*fyne.Container)
			label := itemContainer.Objects[0].(*widget.Label)
			deleteButton := itemContainer.Objects[1].(*widget.Button)

			label.SetText(ui.Servers[i].Name)

			// Capture the current index to avoid closure issues
			index := i
			deleteButton.OnTapped = func() {
				// Confirm deletion (optional)
				confirm := dialog.NewConfirm("Delete Server", "Are you sure you want to delete this server?", func(confirmed bool) {
					if confirmed {
						// Remove from the database
						err := database.DeleteServer(ui.Servers[index].ID)
						if err != nil {
							log.Printf("Error deleting server: %v", err)
							return
						}

						// Remove from the servers slice
						ui.Servers = append(ui.Servers[:index], ui.Servers[index+1:]...)

						// Refresh the list
						ui.ServerList.Refresh()
					}
				}, myWindow)
				confirm.Show()
			}
		},
	)

	ui.ServerList.OnSelected = func(id widget.ListItemID) {
		if id != -1 {
			log.Printf("Server selected: %v", id)
			selectedServer := ui.Servers[id]
			ui.SelectedServerID = selectedServer.ID
			log.Printf("Selected server %v", selectedServer)
			ui.DisplayServerOptions(myWindow, database, selectedServer)
		}
	}

	// Main Content Layout
	mainContent := container.NewHSplit(
		container.NewVBox(ui.ServerList),
		ui.TabContainer,
	)
	mainContent.Offset = 0.2

	// Declare fileMenuButton so it can be accessed inside the closure
	var fileMenuButton *widget.Button

	// Initialize fileMenuButton
	fileMenuButton = widget.NewButton("File", func() {
		// Create a menu with 'Add Server' and 'Exit' options
		fileMenu := fyne.NewMenu("File",
			fyne.NewMenuItem("Add Server", func() {
				ui.AddServer(myWindow, database)
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Exit", func() {
				myApp.Quit()
			}),
		)
		// Show the menu as a pop-up at the button's position
		widget.ShowPopUpMenuAtPosition(fileMenu, myWindow.Canvas(),
			fyne.CurrentApp().Driver().AbsolutePositionForObject(fileMenuButton).
				Add(fyne.NewPos(0, fileMenuButton.Size().Height)))
	})

	// Create the dark mode toggle
	darkModeToggle := widget.NewCheck("Dark Mode", func(checked bool) {
		if checked {
			myApp.Settings().SetTheme(themes.DraculaTheme{})
		} else {
			myApp.Settings().SetTheme(themes.LightTheme{})
		}

		// Save the setting
		err := database.SetSetting("dark_mode", fmt.Sprintf("%v", checked))
		if err != nil {
			log.Printf("Error saving dark mode setting: %v", err)
		}
	})
	darkModeToggle.SetChecked(isDarkMode)

	// Create the top bar with the 'File' button on the left and the dark mode toggle on the right
	topBar := container.NewHBox(
		fileMenuButton,
		layout.NewSpacer(),
		darkModeToggle,
	)

	content := container.NewBorder(topBar, nil, nil, nil, mainContent)
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
