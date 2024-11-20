package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/db"
	"github.com/devalexandre/brokers-ui/state"
)

// AddDashboardTab creates and adds a dashboard tab to the TabContainer
func AddDashboardTab(database *db.Database) {
	// Map to store metrics (labels) for each subscription
	metrics := make(map[string]*widget.Label)

	// Create a container for the dashboard content
	dashboardContainer := container.NewVBox(
		widget.NewLabel("Message Monitoring Dashboard"),
	)

	// Add a section for each subscription with a message counter
	for _, sub := range database.Subs {
		label := widget.NewLabel(fmt.Sprintf("Sub: %s - Messages received: 0", sub.SubName))
		metrics[sub.SubName] = label
		dashboardContainer.Add(label)
	}

	// Create and add the dashboard tab
	dashboardTab := container.NewTabItem("Dashboard", dashboardContainer)
	TabContainer.Append(dashboardTab)
	TabContainer.Select(dashboardTab)

	// Goroutine to monitor messages and update metrics in real-time
	go func() {
		for {
			for subName, label := range metrics {
				// Count the number of messages received for this subscription
				count := len(state.ReceivedMessages[subName])
				label.SetText(fmt.Sprintf("Sub: %s - Messages received: %d", subName, count))
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
