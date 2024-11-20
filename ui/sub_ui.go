package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/db"
	"github.com/devalexandre/brokers-ui/state"
)

func AddSub(window fyne.Window, db *db.Database, serverID int) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter sub name...")

	dialog := dialog.NewForm(
		"Add Sub",
		"Confirm",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Sub Name", entry),
		},
		func(confirmed bool) {
			if confirmed {
				db.SaveSub(serverID, entry.Text)
				db.LoadSubs(serverID)
				RefreshTopicsAndSubs(serverID, db)
				AddTabsForTopicsAndSubs(window, db, serverID)
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func CreateSubTabContent(subName string) fyne.CanvasObject {
	messageChan := make(chan string)
	messageContainer := container.NewVBox()

	log.Printf("Creating sub tab content for %s", subName)

	if BrokerServers[SelectedServerID] == nil {
		log.Printf("Nats server %d is nil", SelectedServerID)
		return container.NewVBox()
	}

	state.ReceivedMessages[subName] = []string{}

	go func() {
		err := BrokerServers[SelectedServerID].Subscribe(subName, func(data []byte) {
			payload := string(data)
			log.Printf("Received message from sub %s: %s", subName, payload)

			// Send the payload to the channel
			messageChan <- payload
		})

		if err != nil {
			log.Printf("Error subscribing to subject %s: %v", subName, err)
			return
		}
	}()

	go func() {
		for payload := range messageChan {
			state.ReceivedMessages[subName] = append(state.ReceivedMessages[subName], payload)

			messageContainer.Add(widget.NewLabel(payload))
			messageContainer.Refresh()
		}
	}()

	return container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Sub: %s", subName)),
		messageContainer,
	)
}

func GetSubNames() []string {
	var names []string
	for _, s := range *Subs {
		names = append(names, s.SubName)
	}
	return names
}
