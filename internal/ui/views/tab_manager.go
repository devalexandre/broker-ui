package views

import (
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/devalexandre/broker-ui/internal/models"
	"github.com/devalexandre/broker-ui/internal/services"
	"github.com/devalexandre/broker-ui/internal/ui/components"
	"github.com/devalexandre/broker-ui/internal/ui/components/charts"
)

type TabManager struct {
	tabContainer   *container.AppTabs
	messageService *services.MessageService
	serverService  *services.ServerService
	window         fyne.Window
	dashboardStop  chan struct{}
	auditStop      chan struct{}
}

type dashboardSubMetric struct {
	bar        *widget.ProgressBar
	countLabel *widget.Label
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

Connect to multiple message brokers and **publish / subscribe** with a unified UI.

## Quick start
1. **Add Server**
2. Select a server from the list
3. Create **Publisher** and **Subscriber** tabs

## Supported brokers
- NATS, RabbitMQ, Google Pub/Sub
- Redis (Pub/Sub), MQTT (QoS 0)
`
	welcomeMessage := widget.NewRichTextFromMarkdown(markdownContent)
	welcomeTab := container.NewTabItem("Welcome", welcomeMessage)
	tm.tabContainer.Append(welcomeTab)
}

// ClearTabs removes all tabs
func (tm *TabManager) ClearTabs() {
	if tm.dashboardStop != nil {
		close(tm.dashboardStop)
		tm.dashboardStop = nil
	}
	if tm.auditStop != nil {
		close(tm.auditStop)
		tm.auditStop = nil
	}
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

	// Add audit tab
	tm.AddAuditTab(currentServer)

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
	if tm.dashboardStop != nil {
		close(tm.dashboardStop)
	}
	tm.dashboardStop = make(chan struct{})

	title := widget.NewLabelWithStyle("Message Monitoring Dashboard", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	totalLabel := widget.NewLabel("Total received: 0")
	rateLabel := widget.NewLabel("Rate: 0 msg/s")

	spark := charts.NewSparkline(60, theme.ForegroundColor(), color.NRGBA{R: 0, G: 0, B: 0, A: 0})

	metrics := make(map[string]*dashboardSubMetric)

	perSubContainer := container.NewVBox()
	for _, sub := range subscriptions {
		bar := widget.NewProgressBar()
		bar.Min = 0
		bar.Max = 1
		bar.SetValue(0)

		count := widget.NewLabel("0")
		metrics[sub.SubName] = &dashboardSubMetric{bar: bar, countLabel: count}

		row := container.NewBorder(
			nil,
			nil,
			widget.NewLabel(sub.SubName),
			count,
			bar,
		)
		perSubContainer.Add(row)
	}

	content := container.NewVBox(
		title,
		container.NewHBox(totalLabel, layout.NewSpacer(), rateLabel),
		widget.NewLabelWithStyle("Msgs/s (últimos 60s)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		spark,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Mensagens por subscription (relativo)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		perSubContainer,
	)

	dashboardTab := container.NewTabItem("Dashboard", container.NewVScroll(content))
	tm.tabContainer.Append(dashboardTab)

	go tm.monitorDashboard(tm.dashboardStop, subscriptions, totalLabel, rateLabel, spark, metrics)
}

// AddAuditTab adds an audit tab (received messages history) for a server.
func (tm *TabManager) AddAuditTab(server models.Server) {
	if tm.auditStop != nil {
		close(tm.auditStop)
	}
	tm.auditStop = make(chan struct{})

	title := widget.NewLabelWithStyle("Audit (Received Messages)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	info := widget.NewLabel("Loading…")

	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter by subject (contains)…")

	limitSelect := widget.NewSelect([]string{"50", "100", "200"}, func(string) {})
	limitSelect.SetSelected("100")

	var refresh func()
	var autoStop chan struct{}
	autoRefresh := widget.NewCheck("Auto-refresh (2s)", func(enabled bool) {
		if autoStop != nil {
			close(autoStop)
			autoStop = nil
		}
		if !enabled {
			return
		}
		autoStop = make(chan struct{})
		go func(appStop <-chan struct{}, localStop <-chan struct{}) {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-appStop:
					return
				case <-localStop:
					return
				case <-ticker.C:
					if refresh != nil {
						refresh()
					}
				}
			}
		}(tm.auditStop, autoStop)
	})

	prevBtn := widget.NewButton("Prev", func() {})
	nextBtn := widget.NewButton("Next", func() {})

	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		if refresh != nil {
			refresh()
		}
	})

	daysEntry := widget.NewEntry()
	daysEntry.SetText("30")
	daysEntry.SetPlaceHolder("Days")
	cleanupBtn := widget.NewButtonWithIcon("Cleanup", theme.DeleteIcon(), func() {})

	var (
		messages []models.Message
		offset   = 0
		total    = -1
	)

	list := widget.NewList(
		func() int { return len(messages) },
		func() fyne.CanvasObject { return widget.NewLabel("…") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(messages) {
				o.(*widget.Label).SetText("")
				return
			}
			msg := messages[i]
			payload := strings.ReplaceAll(msg.Payload, "\n", " ")
			o.(*widget.Label).SetText(fmt.Sprintf("%s | %s | %s", msg.CreatedAt, msg.Subject, truncate(payload, 120)))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(messages) {
			return
		}
		msg := messages[id]
		body := fmt.Sprintf("Subject: %s\nProvider: %s\nAt: %s\n\n%s", msg.Subject, msg.ProviderType, msg.CreatedAt, msg.Payload)
		dialog.ShowInformation("Audit Message", body, tm.window)
		list.Unselect(id)
	}

	refresh = func() {
		limit, _ := strconv.Atoi(limitSelect.Selected)
		if limit <= 0 {
			limit = 100
		}
		filter := strings.TrimSpace(filterEntry.Text)

		go func(curOffset int) {
			var (
				msgs []models.Message
				err  error
			)

			if filter == "" {
				msgs, err = tm.messageService.GetAuditMessages(server.ID, limit, curOffset)
			} else {
				msgs, err = tm.messageService.GetAuditMessagesBySubject(server.ID, filter, limit, curOffset)
			}
			if err != nil {
				fyne.Do(func() { components.ErrorDialog(err, tm.window) })
				return
			}

			t, err := tm.messageService.CountAuditMessages(server.ID)
			if err != nil {
				t = -1
			}

			fyne.Do(func() {
				offset = curOffset
				messages = msgs
				total = t
				list.Refresh()

				from := offset + 1
				to := offset + len(messages)
				if len(messages) == 0 {
					from, to = 0, 0
				}

				totalText := "?"
				if total >= 0 {
					totalText = strconv.Itoa(total)
				}
				info.SetText(fmt.Sprintf("Server: %s | Showing %d–%d of %s", server.Name, from, to, totalText))

				prevBtn.Disable()
				nextBtn.Disable()
				if offset > 0 {
					prevBtn.Enable()
				}
				if len(messages) == limit && (total < 0 || offset+limit < total) {
					nextBtn.Enable()
				}
			})
		}(offset)
	}

	prevBtn.OnTapped = func() {
		limit, _ := strconv.Atoi(limitSelect.Selected)
		if limit <= 0 {
			limit = 100
		}
		newOffset := offset - limit
		if newOffset < 0 {
			newOffset = 0
		}
		offset = newOffset
		refresh()
	}
	nextBtn.OnTapped = func() {
		limit, _ := strconv.Atoi(limitSelect.Selected)
		if limit <= 0 {
			limit = 100
		}
		offset += limit
		refresh()
	}

	filterEntry.OnSubmitted = func(string) {
		offset = 0
		refresh()
	}
	limitSelect.OnChanged = func(string) {
		offset = 0
		refresh()
	}

	cleanupBtn.OnTapped = func() {
		days, err := strconv.Atoi(strings.TrimSpace(daysEntry.Text))
		if err != nil || days < 1 {
			components.ErrorDialog(fmt.Errorf("invalid days value"), tm.window)
			return
		}

		components.ConfirmDialog(
			"Cleanup Audit Messages",
			fmt.Sprintf("Delete audit messages older than %d days for server '%s'?", days, server.Name),
			func(confirmed bool) {
				if !confirmed {
					return
				}
				go func() {
					if err := tm.messageService.CleanupOldAuditMessagesForServer(server.ID, days); err != nil {
						fyne.Do(func() { components.ErrorDialog(err, tm.window) })
						return
					}
					fyne.Do(func() {
						offset = 0
						refresh()
					})
				}()
			},
			tm.window,
		).Show()
	}

	header := container.NewVBox(
		title,
		container.NewHBox(widget.NewLabel("Limit:"), limitSelect, layout.NewSpacer(), autoRefresh),
		container.NewBorder(nil, nil, widget.NewLabel("Subject:"), refreshBtn, filterEntry),
		container.NewHBox(prevBtn, nextBtn, layout.NewSpacer(), widget.NewLabel("Days:"), daysEntry, cleanupBtn),
		info,
		widget.NewSeparator(),
	)

	tab := container.NewTabItem("Audit", container.NewBorder(header, nil, nil, nil, list))
	tm.tabContainer.Append(tab)

	refresh()
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

		err := tm.messageService.PublishMessage(provider, topic.ServerID, subject, payload)
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
		err := tm.messageService.Subscribe(provider, subscription.ServerID, subscription.SubName, subscription.SubjectPattern, messageChan)
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

	urlEntry.OnChanged = func(value string) {
		detected := tm.serverService.DetectProviderFromURL(value)
		providerSelect.SetSelected(string(detected))
	}

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

func (tm *TabManager) monitorDashboard(
	stop <-chan struct{},
	subscriptions []models.Subscription,
	totalLabel *widget.Label,
	rateLabel *widget.Label,
	spark *charts.Sparkline,
	metrics map[string]*dashboardSubMetric,
) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	prevTotal := 0
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			counts := tm.messageService.GetDashboardCounts()

			total := 0
			maxCount := 0
			for _, sub := range subscriptions {
				c := counts[sub.SubName]
				total += c
				if c > maxCount {
					maxCount = c
				}
			}
			if maxCount == 0 {
				maxCount = 1
			}

			rate := total - prevTotal
			prevTotal = total

			fyne.Do(func() {
				totalLabel.SetText("Total received: " + strconv.Itoa(total))
				rateLabel.SetText("Rate: " + strconv.Itoa(rate) + " msg/s")
				spark.Push(float64(rate))

				for _, sub := range subscriptions {
					m := metrics[sub.SubName]
					if m == nil {
						continue
					}
					c := counts[sub.SubName]
					m.countLabel.SetText(strconv.Itoa(c))
					m.bar.SetValue(float64(c) / float64(maxCount))
				}
			})
		}
	}
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return s[:max-1] + "…"
}
