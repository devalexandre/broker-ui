package views

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/devalexandre/broker-ui/internal/models"
	"github.com/devalexandre/broker-ui/internal/services"
	"github.com/devalexandre/broker-ui/internal/ui/components"
)

type TabManager struct {
	tabContainer   *container.AppTabs
	messageService *services.MessageService
	serverService  *services.ServerService
	window         fyne.Window
}

// NewTabManager creates a new tab manager
func NewTabManager(messageService *services.MessageService, serverService *services.ServerService, window fyne.Window) *TabManager {
	return &TabManager{
		tabContainer:   container.NewAppTabs(),
		messageService: messageService,
		serverService:  serverService,
		window:         window,
	}
}

// GetTabContainer returns the tab container
func (tm *TabManager) GetTabContainer() *container.AppTabs {
	return tm.tabContainer
}

// ShowWelcome shows the welcome tab
func (tm *TabManager) ShowWelcome() {
	markdownContent := `# Welcome to Broker UI

This application allows you to connect to **NATS** , **RabbitMQ** , and other message brokers, create topics, and subscribe to subjects.

- **Add Server**: Add a new server connection.
- **Topics**: Publish messages to topics.
- **Subscriptions**: Receive messages from subjects.

Developed by [Alexandre E Souza](https://www.linkedin.com/in/devevantelista)
`
	welcomeMessage := widget.NewRichTextFromMarkdown(markdownContent)
	welcomeTab := container.NewTabItem("Welcome", welcomeMessage)
	tm.tabContainer.Append(welcomeTab)
}

// ClearTabs removes all tabs
func (tm *TabManager) ClearTabs() {
	tm.tabContainer.Items = []*container.TabItem{}
	tm.tabContainer.Refresh()
}

// RefreshServerTabs reloads all tabs for a specific server
func (tm *TabManager) RefreshServerTabs(serverID int) {
	// Get server info
	servers, err := tm.serverService.GetAllServers()
	if err != nil {
		log.Printf("Error getting servers: %v", err)
		return
	}

	var currentServer models.Server
	found := false
	for _, server := range servers {
		if server.ID == serverID {
			currentServer = server
			found = true
			break
		}
	}

	if !found {
		log.Printf("Server with ID %d not found", serverID)
		return
	}

	// Clear existing tabs
	tm.ClearTabs()

	// Add server config tab
	tm.AddServerConfigTab(currentServer)

	// Load and add topic tabs
	topics, err := tm.serverService.GetTopicsForServer(serverID)
	if err != nil {
		log.Printf("Error loading topics: %v", err)
	} else {
		for _, topic := range topics {
			tm.AddTopicTab(topic)
		}
	}

	// Load and add subscription tabs
	subscriptions, err := tm.serverService.GetSubscriptionsForServer(serverID)
	if err != nil {
		log.Printf("Error loading subscriptions: %v", err)
	} else {
		for _, sub := range subscriptions {
			tm.AddSubscriptionTab(sub)
		}
		// Add dashboard tab
		tm.AddDashboardTab(subscriptions)
	}

	tm.tabContainer.Refresh()
	log.Printf("Refreshed tabs for server %d", serverID)
}

// AddServerConfigTab adds a configuration tab for the server
func (tm *TabManager) AddServerConfigTab(server models.Server) {
	menu := components.ServerMenu(
		func() { tm.showAddTopicDialog(server.ID) },
		func() { tm.showAddSubscriptionDialog(server.ID) },
		func() { tm.disconnectServer(server.ID) },
	)

	editButton := widget.NewButtonWithIcon("Edit Connection", theme.ViewRefreshIcon(), func() {
		tm.showEditServerDialog(server)
	})

	panel := container.NewVBox(
		menu,
		widget.NewLabel(fmt.Sprintf("Connected to %s (%s)", server.Name, server.URL)),
		editButton,
	)

	configTab := container.NewTabItem("Config", panel)
	tm.tabContainer.Append(configTab)
	tm.tabContainer.Select(configTab)
}

// AddDashboardTab adds a dashboard monitoring tab
func (tm *TabManager) AddDashboardTab(subscriptions []models.Subscription) {
	dashboardContainer := container.NewVBox(
		widget.NewLabel("Message Monitoring Dashboard"),
	)

	metrics := make(map[string]*widget.Label)
	for _, sub := range subscriptions {
		label := widget.NewLabel(fmt.Sprintf("Sub: %s - Messages received: 0", sub.SubName))
		metrics[sub.SubName] = label
		dashboardContainer.Add(label)
	}

	dashboardTab := container.NewTabItem("Dashboard", dashboardContainer)
	tm.tabContainer.Append(dashboardTab)

	// Start monitoring
	go tm.monitorMessages(metrics)
}

// AddTopicTab adds a tab for publishing messages to a topic
func (tm *TabManager) AddTopicTab(topic models.Topic) {
	messageContainer := container.NewVBox()

	subjectEntry := widget.NewEntry()
	subjectEntry.SetText(topic.TopicName)
	subjectEntry.SetPlaceHolder("Enter subject to publish to...")

	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Enter message payload here...")

	sendButton := widget.NewButton("Send", func() {
		subject := subjectEntry.Text
		payload := messageEntry.Text
		if subject == "" {
			subject = topic.TopicName
		}

		provider, ok := tm.serverService.GetMessagingProvider(topic.ServerID)
		if !ok {
			components.ErrorDialog(fmt.Errorf("no messaging provider connection for server"), tm.window)
			return
		}

		err := tm.messageService.PublishMessage(provider, subject, payload)
		if err != nil {
			components.ErrorDialog(err, tm.window)
			return
		}

		messageContainer.Add(widget.NewLabel(payload))
		messageContainer.Refresh()
		messageEntry.SetText("")
	})

	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		tm.showDeleteTopicDialog(topic)
	})

	content := container.NewVBox(
		container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Publisher: %s", topic.TopicName)),
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

	topicName := fmt.Sprintf("topic-%v", topic.TopicName)
	tab := container.NewTabItemWithIcon(topicName, theme.MailSendIcon(), content)
	tm.tabContainer.Append(tab)
}

// AddSubscriptionTab adds a tab for receiving messages from a subscription
func (tm *TabManager) AddSubscriptionTab(subscription models.Subscription) {
	messageContainer := container.NewVBox()
	messageChan := make(chan string, 100)

	provider, ok := tm.serverService.GetMessagingProvider(subscription.ServerID)
	if !ok {
		log.Printf("No messaging provider connection for server %d", subscription.ServerID)
		return
	}

	// Start subscription
	go func() {
		err := tm.messageService.Subscribe(provider, subscription.SubName, subscription.SubjectPattern, messageChan)
		if err != nil {
			log.Printf("Error subscribing to %s: %v", subscription.SubjectPattern, err)
		}
	}()

	// Monitor messages
	go func() {
		for payload := range messageChan {
			messageContainer.Add(widget.NewLabel(payload))
			messageContainer.Refresh()
		}
	}()

	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		tm.showDeleteSubscriptionDialog(subscription)
	})

	content := container.NewVBox(
		container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Sub: %s (Pattern: %s)", subscription.SubName, subscription.SubjectPattern)),
			closeButton,
		),
		messageContainer,
	)

	subName := fmt.Sprintf("sub-%v", subscription.SubName)
	tab := container.NewTabItemWithIcon(subName, theme.ViewRefreshIcon(), content)
	tm.tabContainer.Append(tab)
}

// Helper methods for dialogs and operations
func (tm *TabManager) showAddTopicDialog(serverID int) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter topic name...")

	dialog := components.FormDialog(
		"Add Topic",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Topic Name", entry),
		},
		func(confirmed bool) {
			if confirmed && entry.Text != "" {
				err := tm.messageService.SaveTopic(serverID, entry.Text)
				if err != nil {
					log.Printf("Error saving topic: %v", err)
					components.ErrorDialog(err, tm.window)
					return
				}
				// Refresh the server tabs to show the new topic
				tm.RefreshServerTabs(serverID)
			}
		},
		tm.window,
	)
	dialog.Show()
}

func (tm *TabManager) showAddSubscriptionDialog(serverID int) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter subscription name...")

	subjectEntry := widget.NewEntry()
	subjectEntry.SetPlaceHolder("Enter subject pattern (e.g., user.*, orders.>, specific.subject)")

	dialog := components.FormDialog(
		"Add Subscription",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Subscription Name", nameEntry),
			widget.NewFormItem("Subject Pattern", subjectEntry),
		},
		func(confirmed bool) {
			if confirmed && nameEntry.Text != "" && subjectEntry.Text != "" {
				err := tm.messageService.SaveSubscription(serverID, nameEntry.Text, subjectEntry.Text)
				if err != nil {
					log.Printf("Error saving subscription: %v", err)
					components.ErrorDialog(err, tm.window)
					return
				}
				// Refresh the server tabs to show the new subscription
				tm.RefreshServerTabs(serverID)
			}
		},
		tm.window,
	)
	dialog.Resize(fyne.NewSize(500, 250))
	dialog.Show()
}

func (tm *TabManager) showEditServerDialog(server models.Server) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(server.Name)
	urlEntry := widget.NewEntry()
	urlEntry.SetText(server.URL)

	// Provider type selection - get supported providers dynamically
	supportedProviders := tm.serverService.GetSupportedProviders()
	providerSelect := widget.NewSelect(supportedProviders, func(value string) {})
	providerSelect.SetSelected(string(server.ProviderType))

	dialog := components.FormDialog(
		"Edit Server Connection",
		"Save",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Server Name", nameEntry),
			widget.NewFormItem("Server URL", urlEntry),
			widget.NewFormItem("Provider Type", providerSelect),
		},
		func(confirmed bool) {
			if confirmed {
				providerType := messaging.ProviderType(providerSelect.Selected)
				tm.serverService.UpdateServer(server.ID, nameEntry.Text, urlEntry.Text, providerType)
				// Refresh would go here
			}
		},
		tm.window,
	)
	dialog.Show()
}

func (tm *TabManager) showDeleteTopicDialog(topic models.Topic) {
	dialog := components.ConfirmDialog(
		"Delete Publisher",
		"Are you sure you want to delete this publisher?",
		func(confirmed bool) {
			if confirmed {
				tm.messageService.DeleteTopic(topic.TopicName, topic.ServerID)
				tm.removeTabByName(fmt.Sprintf("topic-%v", topic.TopicName))
			}
		},
		tm.window,
	)
	dialog.Show()
}

func (tm *TabManager) showDeleteSubscriptionDialog(subscription models.Subscription) {
	dialog := components.ConfirmDialog(
		"Delete Subscription",
		"Are you sure you want to delete this subscription?",
		func(confirmed bool) {
			if confirmed {
				tm.messageService.DeleteSubscription(subscription.SubName, subscription.ServerID)
				tm.removeTabByName(fmt.Sprintf("sub-%v", subscription.SubName))
			}
		},
		tm.window,
	)
	dialog.Show()
}

func (tm *TabManager) disconnectServer(serverID int) {
	tm.serverService.DisconnectFromServer(serverID)
	tm.ClearTabs()
	tm.ShowWelcome()
	tm.tabContainer.Refresh()
}

func (tm *TabManager) removeTabByName(tabName string) {
	for i, tab := range tm.tabContainer.Items {
		if tab.Text == tabName {
			tm.tabContainer.Remove(tab)
			if i < len(tm.tabContainer.Items) {
				tm.tabContainer.Select(tm.tabContainer.Items[i])
			}
			break
		}
	}
	tm.tabContainer.Refresh()
}

func (tm *TabManager) monitorMessages(metrics map[string]*widget.Label) {
	for {
		counts := tm.messageService.GetDashboardCounts()
		for subName, label := range metrics {
			count := counts[subName]
			label.SetText(fmt.Sprintf("Sub: %s - Messages received: %d", subName, count))
		}
		time.Sleep(1 * time.Second)
	}
}
