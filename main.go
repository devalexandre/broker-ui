package main

import (
	"crypto/ed25519"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/broker-ui/icons"
	"github.com/devalexandre/broker-ui/natscli"
	"github.com/devalexandre/broker-ui/themes/dracula"
	"github.com/devalexandre/broker-ui/themes/light"
	"github.com/fynelabs/fyneselfupdate"
	"github.com/fynelabs/selfupdate"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats.go"
)

var db *sql.DB
var serverList *widget.List
var topicsDropdown *widget.Select
var subsDropdown *widget.Select
var tabContainer *container.AppTabs
var selectedServerID int
var selectedServer server
var NatsError error
var wg sync.WaitGroup

type server struct {
	ID   int
	Name string
	URL  string
}

type topic struct {
	ID        int
	ServerID  int
	TopicName string
}

type sub struct {
	ID             int
	ServerID       int
	SubName        string
	SubjectPattern string
}

var servers []server
var topics []topic
var subs []sub
var natsServers = make(map[int]*natscli.Nats)

// A map to store sent messages for each topic and sub
var sentMessages = make(map[string][]string)
var receivedMessages = make(map[string][]string)
var dasboardReceivedMessages = make(map[string]int)

var myWindow fyne.Window
var myApp fyne.App
var isDarkTheme bool = true // Controls whether using dark theme (Dracula) or light theme (Light)
var themeButton *widget.Button

func main() {

	// Initialize SQLite3 database
	var err error
	db, err = sql.Open("sqlite3", "./nats_servers.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, url TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS topics (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, topic_name TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS subs (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, sub_name TEXT, subject_pattern TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	// Migration: add subject_pattern column if it doesn't exist
	_, err = db.Exec(`ALTER TABLE subs ADD COLUMN subject_pattern TEXT DEFAULT ''`)
	if err != nil {
		// Ignore error if column already exists
		log.Printf("Column subject_pattern may already exist: %v", err)
	}

	// Load servers from database
	loadServers()

	myApp = app.New()
	myApp.Settings().SetTheme(dracula.DraculaTheme{})
	myWindow = myApp.NewWindow("NATS Client")

	// Create theme button
	themeButton = widget.NewButtonWithIcon("Theme", icons.ThemeToggleIcon(isDarkTheme), func() {
		toggleTheme()
	})

	// Top Menu
	menu := container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewButtonWithIcon("Add Server", icons.AddServerIcon(), func() {
				addServer(myWindow)
			}),
			themeButton,
		),
		widget.NewButtonWithIcon("Exit", icons.ExitIcon(), func() {
			myApp.Quit()
		}),
	)

	// Markdown content for the welcome message
	markdownContent := `# Welcome to NATS Client

This application allows you to connect to **NATS** servers, create topics, and subscribe to subjects.

- **Add Server**: Add a new NATS server connection.
- **Topics**: Publish messages to topics.
- **Subscriptions**: Receive messages from subjects.

Developed by [Alexandre E Souza](https://www.linkedin.com/in/devevantelista)
`

	// Create a RichText widget to render the markdown
	welcomeMessage := widget.NewRichTextFromMarkdown(markdownContent)

	// Main Panel
	tabContainer = container.NewAppTabs()
	welcomeTab := container.NewTabItem("Welcome", welcomeMessage)
	tabContainer.Append(welcomeTab)
	// Server List
	serverList = widget.NewList(
		func() int { return len(servers) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Server Name")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(servers[i].Name)
		},
	)

	serverList.OnSelected = func(id widget.ListItemID) {
		selectedServer = servers[id]
		selectedServerID = selectedServer.ID
		displayServerOptions(myWindow, selectedServer.Name, selectedServer.URL)
	}

	// Main Layout
	mainContent := container.NewHSplit(
		container.NewVBox(serverList),
		tabContainer,
	)
	mainContent.Offset = 0.2 // Define left menu space

	content := container.NewBorder(menu, nil, nil, nil, mainContent)

	//update
	selfManage(myApp, myWindow)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()

}

func displayServerOptions(window fyne.Window, name string, url string) {
	// Clear tabs before recreating
	clearTabs()

	// Additional menu for selected server
	menu := container.NewHBox(
		widget.NewButtonWithIcon("Add Publisher", icons.PublisherIcon(), func() {
			addTopic(window, selectedServerID)
		}),
		widget.NewButtonWithIcon("Add Subscriber", icons.SubscriberIcon(), func() {
			addSub(window, selectedServerID)
		}),
		widget.NewButtonWithIcon("Disconnect", theme.MediaStopIcon(), func() {
			disconnectFromServer()
		}),
	)

	// Edit Connection Button
	editButton := widget.NewButtonWithIcon("Edit Connection", theme.ViewRefreshIcon(), func() {
		editServerConnection(window, selectedServerID, name, url)
	})

	panel := container.NewVBox(
		menu,
		widget.NewLabel(fmt.Sprintf("Connected to %s (%s)", name, url)),
		editButton,
	)

	// Attempt to create a new NATS client
	natsServers[selectedServerID], NatsError = natscli.NewNats(url)
	if NatsError != nil {
		dialog.ShowError(NatsError, window)
	} else {
		// Load topics and subs if connection is successful
		loadTopics(selectedServerID)
		loadSubs(selectedServerID)
	}

	configTab := container.NewTabItem("Config", panel)
	tabContainer.Append(configTab)
	tabContainer.Select(configTab)
	tabContainer.Refresh()

	addDashboardTab()
	addTabsForTopicsAndSubs(selectedServerID)
}

func disconnectFromServer() {
	// Close NATS connection and clear tabs
	if client, ok := natsServers[selectedServerID]; ok {
		client.Close()
		delete(natsServers, selectedServerID)
	}
	clearTabs()
	tabContainer.Append(container.NewTabItem("Welcome", widget.NewLabel("Welcome to the NATS Client!")))
	tabContainer.Refresh()
}

func clearTabs() {
	tabContainer.Items = []*container.TabItem{}
	tabContainer.Refresh()
}

func editServerConnection(window fyne.Window, serverID int, currentName, currentURL string) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(currentName)
	urlEntry := widget.NewEntry()
	urlEntry.SetText(currentURL)

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
				updateServer(serverID, nameEntry.Text, urlEntry.Text)
				loadServers()        // Reload server list
				serverList.Refresh() // Update displayed list
				displayServerOptions(window, nameEntry.Text, urlEntry.Text)
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func updateServer(serverID int, name string, url string) {
	stmt, err := db.Prepare("UPDATE servers SET name = ?, url = ? WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, url, serverID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server updated:", name, url)
}

func addTabsForTopicsAndSubs(serverID int) {
	for _, t := range topics {
		topicName := fmt.Sprintf("topic-%v", t.TopicName)
		tab := container.NewTabItemWithIcon(topicName, theme.MailSendIcon(), createTopicTabContent(t.TopicName, serverID))
		tabContainer.Append(tab)
	}

	for _, s := range subs {
		subname := fmt.Sprintf("sub-%v", s.SubName)
		tab := container.NewTabItemWithIcon(subname, theme.ViewRefreshIcon(), createSubTabContent(s.SubName, serverID))
		tabContainer.Append(tab)
	}

	tabContainer.Refresh()
}

func createSubTabContent(subName string, serverId int) fyne.CanvasObject {
	// Find corresponding subscription to get subject pattern
	var subjectPattern string
	for _, s := range subs {
		if s.SubName == subName && s.ServerID == serverId {
			subjectPattern = s.SubjectPattern
			break
		}
	}

	// If pattern not found, use sub name as fallback
	if subjectPattern == "" {
		subjectPattern = subName
	}

	// Channel to receive messages
	messageChan := make(chan string)

	// Vertical box to store messages
	messageContainer := container.NewVBox()

	receivedMessages[subName] = []string{} // Initialize received messages list for sub

	// Goroutine to handle NATS subscription and send messages to channel
	go func() {
		err := natsServers[selectedServerID].Subscribe(subjectPattern, func(m *nats.Msg) {
			payload := string(m.Data)
			subject := m.Subject
			log.Printf("Received message from sub %s (subject: %s): %s", subName, subject, payload)

			// Send message to channel with subject information
			messageChan <- fmt.Sprintf("[%s] %s", subject, payload)
		})

		if err != nil {
			log.Printf("Error subscribing to subject pattern %s: %v", subjectPattern, err)
			return
		}
	}()

	// Goroutine to monitor channel and update GUI
	go func() {
		for payload := range messageChan {
			receivedMessages[subName] = append(receivedMessages[subName], payload)

			// Add message to container in main thread
			messageContainer.Add(widget.NewLabel(payload))

			// Update GUI
			messageContainer.Refresh()
		}
	}()

	// Create "X" button to close tab
	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		dialog.NewConfirm("Delete Subscription", "Are you sure you want to delete this subscription?", func(confirmed bool) {
			if confirmed {
				deleteSub(subName, serverId)
				tab := findTabBySubName(subName)
				tabContainer.Remove(tab)
				tabContainer.Refresh()
			}
		}, myWindow).Show()
	})

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Sub: %s (Pattern: %s)", subName, subjectPattern)),
			closeButton,
		),
		messageContainer,
	)
}

func createTopicTabContent(topicName string, serverId int) fyne.CanvasObject {
	// Slice to store messages
	var messages []string

	// Vertical container to display messages
	messageContainer := container.NewVBox()

	// Text entry for subject
	subjectEntry := widget.NewEntry()
	subjectEntry.SetText(topicName) // Pre-fill with topic name
	subjectEntry.SetPlaceHolder("Enter subject to publish to...")

	// Text entry for message
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Enter message payload here...")

	// Send button
	sendButton := widget.NewButton("Send", func() {
		subject := subjectEntry.Text
		payload := messageEntry.Text
		if subject == "" {
			subject = topicName // Fallback to topic name
		}
		sendMessageToTopic(subject, payload, messageContainer, &messages)

		// Clear message entry
		messageEntry.SetText("")
	})

	// Create "X" button to close tab
	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		dialog.NewConfirm("Delete Publisher", "Are you sure you want to delete this publisher?", func(confirmed bool) {
			if confirmed {
				deleteTopic(topicName, serverId)
				tab := findTabByTopicName(topicName)
				tabContainer.Remove(tab)
				tabContainer.Refresh()
			}
		}, myWindow).Show()

	})

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Publisher: %s", topicName)),
			closeButton,
		),
		widget.NewLabel("Subject:"),
		subjectEntry,
		widget.NewLabel("Message:"),
		messageEntry,
		sendButton,
		widget.NewSeparator(),
		widget.NewLabel("Sent Messages:"),
		messageContainer,
	)
}

func sendMessageToTopic(topicName string, payload string, messageContainer *fyne.Container, messages *[]string) {
	if payload == "" {
		return
	}

	// Send message to NATS server
	err := natsServers[selectedServerID].Publish(topicName, []byte(payload))
	if err != nil {
		fmt.Println("Error publishing message:", err)
		return
	}

	log.Printf("Sending message to topic %s: %s", topicName, payload)

	// Store message in list for display
	*messages = append(*messages, payload)

	// Update GUI
	messageContainer.Add(widget.NewLabel(payload))
	messageContainer.Refresh()
}

func findTabByTopicName(topicName string) *container.TabItem {
	for _, tab := range tabContainer.Items {
		if tab.Text == fmt.Sprintf("topic-%v", topicName) {
			return tab
		}
	}
	return nil
}

func findTabBySubName(subName string) *container.TabItem {
	for _, tab := range tabContainer.Items {
		if tab.Text == fmt.Sprintf("sub-%v", subName) {
			return tab
		}
	}
	return nil
}

func addServer(parent fyne.Window) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter server name...")
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter server URL...")

	dialog := dialog.NewForm(
		"Add Server",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Server Name", nameEntry),
			widget.NewFormItem("Server URL", urlEntry),
		},
		func(confirmed bool) {
			if confirmed {
				saveServer(nameEntry.Text, urlEntry.Text)
				displayServerOptions(myWindow, selectedServer.Name, selectedServer.URL)
			}
		},
		parent,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func saveServer(name string, url string) {
	if name == "" || url == "" {
		return
	}

	// Insert Name and URL into SQLite3 database
	stmt, err := db.Prepare("INSERT INTO servers(name, url) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, url)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server saved:", name, url)
}

func loadServers() {
	// Load servers from database
	rows, err := db.Query("SELECT id, name, url FROM servers")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	servers = []server{}
	for rows.Next() {
		var s server
		err := rows.Scan(&s.ID, &s.Name, &s.URL)
		if err != nil {
			log.Fatal(err)
		}
		servers = append(servers, s)
	}
}

func addTopic(window fyne.Window, serverID int) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter topic name...")

	dialog := dialog.NewForm(
		"Add Topic",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Topic Name", entry),
		},
		func(confirmed bool) {
			if confirmed {
				saveTopic(serverID, entry.Text)
				displayServerOptions(myWindow, selectedServer.Name, selectedServer.URL)
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func saveTopic(serverID int, topicName string) {
	if topicName == "" {
		return
	}
	q := "INSERT INTO topics(server_id, topic_name) VALUES(?, ?)"
	stmt, err := db.Prepare(q)
	if err != nil {
		log.Printf("saveTopic query %v", q)
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID, topicName)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Topic saved:", topicName)
}

func loadTopics(serverID int) {
	rows, err := db.Query("SELECT id, topic_name FROM topics WHERE server_id = ?", serverID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	topics = []topic{}
	for rows.Next() {
		var t topic
		err := rows.Scan(&t.ID, &t.TopicName)
		if err != nil {
			log.Fatal(err)
		}
		topics = append(topics, t)
	}
}

func addSub(window fyne.Window, serverID int) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter subscription name...")

	subjectEntry := widget.NewEntry()
	subjectEntry.SetPlaceHolder("Enter subject pattern (e.g., user.*, orders.>, specific.subject)")

	dialog := dialog.NewForm(
		"Add Subscription",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Subscription Name", nameEntry),
			widget.NewFormItem("Subject Pattern", subjectEntry),
		},
		func(confirmed bool) {
			if confirmed {
				saveSub(serverID, nameEntry.Text, subjectEntry.Text)
				displayServerOptions(myWindow, selectedServer.Name, selectedServer.URL)
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(500, 250))
	dialog.Show()
}

func saveSub(serverID int, subName string, subjectPattern string) {
	if subName == "" || subjectPattern == "" {
		return
	}

	stmt, err := db.Prepare("INSERT INTO subs(server_id, sub_name, subject_pattern) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID, subName, subjectPattern)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Sub saved: %s (pattern: %s)", subName, subjectPattern)
}

func loadSubs(serverID int) {
	rows, err := db.Query("SELECT id, sub_name, COALESCE(subject_pattern, '') FROM subs WHERE server_id = ?", serverID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	subs = []sub{}
	for rows.Next() {
		var s sub
		err := rows.Scan(&s.ID, &s.SubName, &s.SubjectPattern)
		if err != nil {
			log.Fatal(err)
		}
		s.ServerID = serverID
		subs = append(subs, s)
	}
}

func deleteTopic(topicname string, serverId int) {
	stmt, err := db.Prepare("DELETE FROM topics WHERE topic_name = ? AND server_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(topicname, serverId)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Topic deleted:", topicname)
}

func deleteSub(subName string, serverId int) {
	stmt, err := db.Prepare("DELETE FROM subs WHERE sub_name = ? AND server_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(subName, serverId)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Sub deleted:", subName)
}

func getTopicNames() []string {
	var names []string
	for _, t := range topics {
		names = append(names, t.TopicName)
	}
	return names
}

func getSubNames() []string {
	var names []string
	for _, s := range subs {
		names = append(names, s.SubName)
	}
	return names
}

// Add Dashboard tab
func addDashboardTab() {
	// Variables to store message counts per sub
	metrics := make(map[string]*widget.Label)

	// Create container for charts or tables for each sub
	dashboardContainer := container.NewVBox(
		widget.NewLabel("Message Monitoring Dashboard"),
	)

	// Add section for each sub with message counter
	for _, sub := range subs {
		label := widget.NewLabel(fmt.Sprintf("Sub: %s - Messages received: 0", sub.SubName))
		metrics[sub.SubName] = label
		dashboardContainer.Add(label)
	}

	// Add Dashboard tab to tabContainer
	dashboardTab := container.NewTabItem("Dashboard", dashboardContainer)
	tabContainer.Append(dashboardTab)
	tabContainer.Select(dashboardTab)

	// Goroutine to monitor messages and update metrics
	go func() {
		for {
			for subName, label := range metrics {
				// Count number of messages received for this sub
				count := len(receivedMessages[subName])
				label.SetText(fmt.Sprintf("Sub: %s - Messages received: %d", subName, count))
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

// toggleTheme switches between dark theme (Dracula) and light theme (Light)
func toggleTheme() {
	if isDarkTheme {
		// Switch to light theme
		myApp.Settings().SetTheme(light.LightTheme{})
		isDarkTheme = false
		log.Println("Switched to Light theme")
	} else {
		// Switch to dark theme
		myApp.Settings().SetTheme(dracula.DraculaTheme{})
		isDarkTheme = true
		log.Println("Switched to Dracula theme")
	}

	// Update theme button icon
	themeButton.SetIcon(icons.ThemeToggleIcon(isDarkTheme))
}

func selfManage(a fyne.App, w fyne.Window) {
	publicKey := ed25519.PublicKey{226, 162, 120, 210, 212, 122, 98, 250, 123, 180, 135, 69, 168, 77, 125, 41, 229, 245, 5, 32, 82, 254, 3, 37, 24, 224, 244, 63, 161, 123, 212, 197}

	// The public key above matches the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "https://geoffrey-artefacts.fynelabs.com/self-update/5d/5de43cb1-be73-4588-9c48-b2acb6169de0/{{.OS}}-{{.Arch}}/{{.Executable}}{{.Ext}}")

	config := fyneselfupdate.NewConfigWithTimeout(a, w, time.Minute, httpSource, selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * 12}, publicKey)

	_, err := selfupdate.Manage(config)
	if err != nil {
		fyne.LogError("Failed to set up update manager", err)
	}
}
