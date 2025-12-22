package views

import (
	"image/color"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/broker-ui/icons"
	"github.com/devalexandre/broker-ui/internal/database"
	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/devalexandre/broker-ui/internal/models"
	"github.com/devalexandre/broker-ui/internal/services"
	"github.com/devalexandre/broker-ui/internal/ui/components"
	"github.com/devalexandre/broker-ui/themes/material"
)

type MainWindow struct {
	app            fyne.App
	window         fyne.Window
	serverService  *services.ServerService
	messageService *services.MessageService
	tabManager     *TabManager
	serverList     *widget.List
	serverFilter   *widget.Entry
	statusLabel    *widget.Label
	isDarkTheme    bool
	themeButton    *widget.Button
	servers        []models.Server
	filtered       []models.Server
	selectedServer models.Server
	hasSelection   bool
}

// NewMainWindow creates a new main window
func NewMainWindow(db *database.Database) *MainWindow {
	// Initialize repositories
	serverRepo := database.NewServerRepository(db.GetDB())
	topicRepo := database.NewTopicRepository(db.GetDB())
	subscriptionRepo := database.NewSubscriptionRepository(db.GetDB())
	messageRepo := database.NewMessageRepository(db.GetDB())

	// Initialize services
	serverService := services.NewServerService(serverRepo, topicRepo, subscriptionRepo)
	messageService := services.NewMessageService(topicRepo, subscriptionRepo, messageRepo)

	// Create Fyne app
	myApp := app.NewWithID("github.com.devalexandre.broker-ui")
	myWindow := myApp.NewWindow("Broker UI")

	isDark := myApp.Preferences().String("theme") != "light"
	if isDark {
		myApp.Settings().SetTheme(material.MaterialTheme{Variant: theme.VariantDark})
	} else {
		myApp.Settings().SetTheme(material.MaterialTheme{Variant: theme.VariantLight})
	}

	mw := &MainWindow{
		app:            myApp,
		window:         myWindow,
		serverService:  serverService,
		messageService: messageService,
		isDarkTheme:    isDark,
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
	// Top AppBar
	appTitle := widget.NewLabelWithStyle("Broker UI", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	mw.themeButton = widget.NewButtonWithIcon("", getThemeIcon(mw.isDarkTheme), mw.toggleTheme)
	mw.themeButton.Importance = widget.LowImportance
	exitButton := widget.NewButtonWithIcon("", icons.ExitIcon(), mw.app.Quit)
	exitButton.Importance = widget.LowImportance

	appBar := container.NewPadded(container.NewHBox(
		appTitle,
		layout.NewSpacer(),
		mw.themeButton,
		exitButton,
	))
	appBarWithDivider := container.NewVBox(appBar, widget.NewSeparator())

	// Create server list
	mw.serverFilter = widget.NewEntry()
	mw.serverFilter.SetPlaceHolder("Search servers…")
	mw.serverFilter.OnChanged = func(string) {
		mw.applyServerFilter()
	}

	mw.serverList = widget.NewList(
		func() int {
			return len(mw.filtered)
		},
		func() fyne.CanvasObject {
			statusDot := canvas.NewCircle(color.NRGBA{R: 0x9A, G: 0xA0, B: 0xA6, A: 0xFF})
			statusDot.Resize(fyne.NewSize(10, 10))

			nameLabel := widget.NewLabelWithStyle("Server Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			metaLabel := widget.NewLabel("Provider • URL")
			metaLabel.Wrapping = fyne.TextTruncate

			deleteBtn := widget.NewButtonWithIcon("", icons.TrashBinIcon(), func() {})
			deleteBtn.Importance = widget.LowImportance

			left := container.NewHBox(statusDot)
			center := container.NewVBox(nameLabel, metaLabel)
			return container.NewBorder(nil, nil, left, deleteBtn, center)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(mw.filtered) {
				return
			}
			server := mw.filtered[i]

			border := o.(*fyne.Container)
			center := border.Objects[0].(*fyne.Container)
			left := border.Objects[1].(*fyne.Container)
			deleteBtn := border.Objects[2].(*widget.Button)

			statusDot := left.Objects[0].(*canvas.Circle)
			nameLabel := center.Objects[0].(*widget.Label)
			metaLabel := center.Objects[1].(*widget.Label)

			nameLabel.SetText(server.Name)
			metaLabel.SetText(string(server.ProviderType) + " • " + server.URL)

			if mw.serverService.IsConnected(server.ID) {
				statusDot.FillColor = color.NRGBA{R: 0x2E, G: 0xCC, B: 0x71, A: 0xFF}
			} else {
				statusDot.FillColor = color.NRGBA{R: 0x9A, G: 0xA0, B: 0xA6, A: 0xFF}
			}
			statusDot.Refresh()

			deleteBtn.OnTapped = func() { mw.showDeleteServerConfirmation(server) }
		},
	)

	mw.serverList.OnSelected = func(id widget.ListItemID) {
		if id < len(mw.filtered) {
			mw.selectServer(mw.filtered[id])
		}
	}

	// Show welcome tab initially
	mw.tabManager.ShowWelcome()

	serversTitle := widget.NewLabelWithStyle("Servers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addServerBtn := widget.NewButtonWithIcon("Add Server", icons.AddServerIcon(), mw.showAddServerDialog)

	sidebarHeader := container.NewVBox(serversTitle, mw.serverFilter)
	sidebar := container.NewBorder(sidebarHeader, addServerBtn, nil, nil, mw.serverList)
	sidebarPadded := container.NewPadded(sidebar)

	mw.statusLabel = widget.NewLabel("Ready")
	statusBar := container.NewVBox(
		widget.NewSeparator(),
		container.NewPadded(container.NewHBox(mw.statusLabel, layout.NewSpacer())),
	)

	mainContent := container.NewHSplit(sidebarPadded, mw.tabManager.GetTabContainer())
	mainContent.Offset = 0.30

	content := container.NewBorder(appBarWithDivider, statusBar, nil, nil, mainContent)
	mw.window.SetContent(content)
	mw.window.Resize(fyne.NewSize(900, 600))

	go mw.periodicUIRefresh()
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

	mw.applyServerFilter()

	log.Printf("Server list refreshed with %d items", len(mw.servers))
}

func (mw *MainWindow) applyServerFilter() {
	if mw.serverFilter == nil || mw.serverList == nil {
		return
	}

	filter := strings.ToLower(strings.TrimSpace(mw.serverFilter.Text))
	if filter == "" {
		mw.filtered = mw.servers
		mw.serverList.Refresh()
		return
	}

	out := make([]models.Server, 0, len(mw.servers))
	for _, s := range mw.servers {
		hay := strings.ToLower(s.Name + " " + s.URL + " " + string(s.ProviderType))
		if strings.Contains(hay, filter) {
			out = append(out, s)
		}
	}
	mw.filtered = out
	mw.serverList.Refresh()
}

func (mw *MainWindow) periodicUIRefresh() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fyne.Do(func() {
			if mw.serverList != nil {
				mw.serverList.Refresh()
			}
			if mw.statusLabel != nil && mw.hasSelection {
				if mw.serverService.IsConnected(mw.selectedServer.ID) {
					mw.statusLabel.SetText("Connected: " + mw.selectedServer.Name + " (" + string(mw.selectedServer.ProviderType) + ")")
				} else {
					mw.statusLabel.SetText("Disconnected: " + mw.selectedServer.Name)
				}
			}
		})
	}
}

// selectServer handles server selection
func (mw *MainWindow) selectServer(server models.Server) {
	// Clear existing tabs
	mw.tabManager.ClearTabs()

	// Connect to server
	err := mw.serverService.ConnectToServer(server.ID, server.URL, server.ProviderType)
	if err != nil {
		components.ErrorDialog(err, mw.window)
		return
	}

	mw.selectedServer = server
	mw.hasSelection = true
	mw.statusLabel.SetText("Connected: " + server.Name + " (" + string(server.ProviderType) + ")")

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

	// Add audit tab (received message history)
	mw.tabManager.AddAuditTab(server)

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

	urlEntry.OnChanged = func(value string) {
		detected := mw.serverService.DetectProviderFromURL(value)
		providerSelect.SetSelected(string(detected))
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
		mw.app.Settings().SetTheme(material.MaterialTheme{Variant: theme.VariantLight})
		mw.isDarkTheme = false
		mw.app.Preferences().SetString("theme", "light")
		log.Println("Switched to Material Light theme")
	} else {
		mw.app.Settings().SetTheme(material.MaterialTheme{Variant: theme.VariantDark})
		mw.isDarkTheme = true
		mw.app.Preferences().SetString("theme", "dark")
		log.Println("Switched to Material Dark theme")
	}

	if mw.themeButton != nil {
		mw.themeButton.SetIcon(getThemeIcon(mw.isDarkTheme))
	}
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
				if mw.hasSelection && mw.selectedServer.ID == server.ID {
					mw.hasSelection = false
					if mw.statusLabel != nil {
						mw.statusLabel.SetText("Ready")
					}
					if mw.serverList != nil {
						mw.serverList.UnselectAll()
					}
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
