package views

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/broker-ui/icons"
	"github.com/devalexandre/broker-ui/internal/database"
	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/devalexandre/broker-ui/internal/models"
	"github.com/devalexandre/broker-ui/internal/services"
	"github.com/devalexandre/broker-ui/internal/ui/components"
	"github.com/devalexandre/broker-ui/themes/dracula"
	"github.com/devalexandre/broker-ui/themes/light"
)

type MainWindow struct {
	app            fyne.App
	window         fyne.Window
	serverService  *services.ServerService
	messageService *services.MessageService
	tabManager     *TabManager
	serverList     *widget.List
	isDarkTheme    bool
	themeButton    *widget.Button
	servers        []models.Server
}

// NewMainWindow creates a new main window
func NewMainWindow(db *database.Database) *MainWindow {
	// Initialize repositories
	serverRepo := database.NewServerRepository(db.GetDB())
	topicRepo := database.NewTopicRepository(db.GetDB())
	subscriptionRepo := database.NewSubscriptionRepository(db.GetDB())

	// Initialize services
	serverService := services.NewServerService(serverRepo, topicRepo, subscriptionRepo)
	messageService := services.NewMessageService(topicRepo, subscriptionRepo)

	// Create Fyne app
	myApp := app.New()
	myApp.Settings().SetTheme(dracula.DraculaTheme{})
	myWindow := myApp.NewWindow("Broker UI")

	mw := &MainWindow{
		app:            myApp,
		window:         myWindow,
		serverService:  serverService,
		messageService: messageService,
		isDarkTheme:    true,
	}

	// Initialize tab manager
	mw.tabManager = NewTabManager(messageService, serverService, myWindow)

	// Setup UI
	mw.setupUI()
	mw.loadServers()

	return mw
}

// setupUI sets up the user interface
func (mw *MainWindow) setupUI() {
	// Create theme button
	mw.themeButton = widget.NewButtonWithIcon("Theme", getThemeIcon(mw.isDarkTheme), mw.toggleTheme)

	// Create main menu
	menu := components.MainMenu(
		mw.showAddServerDialog,
		mw.toggleTheme,
		mw.app.Quit,
		mw.isDarkTheme,
	)

	// Create server list
	mw.serverList = widget.NewList(
		func() int {
			return len(mw.servers)
		},
		func() fyne.CanvasObject {
			// Create a container with server name and delete button
			nameLabel := widget.NewLabel("Server Name")
			deleteBtn := widget.NewButtonWithIcon("", icons.TrashBinIcon(), func() {})
			deleteBtn.Resize(fyne.NewSize(24, 24)) // Small delete button

			return container.NewBorder(nil, nil, nil, deleteBtn, nameLabel)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(mw.servers) {
				server := mw.servers[i]
				borderContainer := o.(*fyne.Container)

				// Update the label (first object in the container)
				nameLabel := borderContainer.Objects[0].(*widget.Label)
				nameLabel.SetText(server.Name)

				// Update the delete button action (second object in the container)
				deleteBtn := borderContainer.Objects[1].(*widget.Button)
				deleteBtn.OnTapped = func() {
					mw.showDeleteServerConfirmation(server)
				}
			}
		},
	)

	mw.serverList.OnSelected = func(id widget.ListItemID) {
		if id < len(mw.servers) {
			mw.selectServer(mw.servers[id])
		}
	}

	// Show welcome tab initially
	mw.tabManager.ShowWelcome()

	// Main layout
	// Create a border container for the server list with a header
	serversHeader := widget.NewLabel("Servers")
	serversHeader.TextStyle = fyne.TextStyle{Bold: true}

	serverListContainer := container.NewBorder(
		serversHeader, // top
		nil,           // bottom
		nil,           // left
		nil,           // right
		mw.serverList, // center - this will take all remaining space
	)

	mainContent := container.NewHSplit(
		serverListContainer,
		mw.tabManager.GetTabContainer(),
	)
	mainContent.Offset = 0.3 // Increased from 0.25 to 0.3 (20% wider)

	content := container.NewBorder(menu, nil, nil, nil, mainContent)
	mw.window.SetContent(content)
	mw.window.Resize(fyne.NewSize(900, 600))
}

// Run starts the application
func (mw *MainWindow) Run() {
	mw.window.ShowAndRun()
}

// GetApp returns the Fyne app
func (mw *MainWindow) GetApp() fyne.App {
	return mw.app
}

// GetWindow returns the main window
func (mw *MainWindow) GetWindow() fyne.Window {
	return mw.window
}

// loadServers loads servers from database and updates the list
func (mw *MainWindow) loadServers() {
	servers, err := mw.serverService.GetAllServers()
	if err != nil {
		log.Printf("Error loading servers: %v", err)
		return
	}

	log.Printf("Loaded %d servers from database", len(servers))
	mw.servers = servers

	// Force the list to rebuild completely
	mw.serverList.Refresh()

	log.Printf("Server list refreshed with %d items", len(mw.servers))
} // selectServer handles server selection
func (mw *MainWindow) selectServer(server models.Server) {
	// Clear existing tabs
	mw.tabManager.ClearTabs()

	// Connect to server
	err := mw.serverService.ConnectToServer(server.ID, server.URL, server.ProviderType)
	if err != nil {
		components.ErrorDialog(err, mw.window)
		return
	}

	// Add server config tab
	mw.tabManager.AddServerConfigTab(server)

	// Load and add topic tabs
	topics, err := mw.serverService.GetTopicsForServer(server.ID)
	if err != nil {
		log.Printf("Error loading topics: %v", err)
	} else {
		for _, topic := range topics {
			mw.tabManager.AddTopicTab(topic)
		}
	}

	// Load and add subscription tabs
	subscriptions, err := mw.serverService.GetSubscriptionsForServer(server.ID)
	if err != nil {
		log.Printf("Error loading subscriptions: %v", err)
	} else {
		for _, sub := range subscriptions {
			mw.tabManager.AddSubscriptionTab(sub)
		}
		// Add dashboard tab
		mw.tabManager.AddDashboardTab(subscriptions)
	}

	mw.tabManager.GetTabContainer().Refresh()
}

// showAddServerDialog shows the dialog to add a new server
func (mw *MainWindow) showAddServerDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter server name...")
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter server URL...")

	// Provider type selection - get supported providers dynamically
	supportedProviders := mw.serverService.GetSupportedProviders()
	providerSelect := widget.NewSelect(supportedProviders, func(value string) {})
	if len(supportedProviders) > 0 {
		providerSelect.SetSelected(supportedProviders[0])
	}

	dialog := components.FormDialog(
		"Add Server",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Server Name", nameEntry),
			widget.NewFormItem("Server URL", urlEntry),
			widget.NewFormItem("Provider Type", providerSelect),
		},
		func(confirmed bool) {
			if confirmed && nameEntry.Text != "" && urlEntry.Text != "" {
				providerType := messaging.ProviderType(providerSelect.Selected)
				err := mw.serverService.SaveServer(nameEntry.Text, urlEntry.Text, providerType)
				if err != nil {
					components.ErrorDialog(err, mw.window)
					return
				}
				mw.loadServers()
			}
		},
		mw.window,
	)
	dialog.Show()
}

// toggleTheme switches between dark and light themes
func (mw *MainWindow) toggleTheme() {
	if mw.isDarkTheme {
		mw.app.Settings().SetTheme(light.LightTheme{})
		mw.isDarkTheme = false
		log.Println("Switched to Light theme")
	} else {
		mw.app.Settings().SetTheme(dracula.DraculaTheme{})
		mw.isDarkTheme = true
		log.Println("Switched to Dracula theme")
	}

	mw.themeButton.SetIcon(getThemeIcon(mw.isDarkTheme))
}

// getThemeIcon returns the appropriate theme icon
func getThemeIcon(isDarkTheme bool) fyne.Resource {
	return icons.ThemeToggleIcon(isDarkTheme)
}

// showDeleteServerConfirmation shows a confirmation dialog before deleting a server
func (mw *MainWindow) showDeleteServerConfirmation(server models.Server) {
	dialog := components.ConfirmDialog(
		"Delete Server",
		"Are you sure you want to delete the server '"+server.Name+"'?\n\nThis action cannot be undone.",
		func(confirmed bool) {
			if confirmed {
				err := mw.serverService.DeleteServer(server.ID)
				if err != nil {
					components.ErrorDialog(err, mw.window)
					return
				}
				// Reload the server list
				mw.loadServers()
				// Clear tabs if this was the selected server
				mw.tabManager.ClearTabs()
				mw.tabManager.ShowWelcome()
			}
		},
		mw.window,
	)
	dialog.Show()
}
