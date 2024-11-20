package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/db"

	"log"
)

func AddTopic(window fyne.Window, db *db.Database, serverID int) {
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
				db.SaveTopic(serverID, entry.Text)
				db.LoadTopics(serverID)
				TopicsDropdown.Options = GetTopicNames(db)
				TopicsDropdown.Refresh()
				AddTabsForTopicsAndSubs(db)
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func CreateTopicTabContent(topicName string) fyne.CanvasObject {
	var messages []string

	messageContainer := container.NewVBox()

	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Enter message payload here...")

	sendButton := widget.NewButton("Send", func() {
		payload := messageEntry.Text
		SendMessageToTopic(topicName, payload, messageContainer, &messages)

		messageEntry.SetText("")
	})

	return container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Topic: %s", topicName)),
		messageEntry,
		sendButton,
		messageContainer,
	)
}

func SendMessageToTopic(topicName, payload string, messageContainer *fyne.Container, messages *[]string) {
	if payload == "" {
		return
	}

	if BrokerServers[SelectedServerID] == nil {
		log.Printf("Nats server %d is nil", SelectedServerID)
		return
	}

	log.Printf("Publishing message to topic %s: %s", topicName, payload)

	err := BrokerServers[SelectedServerID].Publish(topicName, []byte(payload))
	if err != nil {
		log.Println("Error publishing message:", err)
		return
	}

	log.Printf("Sending message to topic %s: %s", topicName, payload)

	*messages = append(*messages, payload)

	messageContainer.Add(widget.NewLabel(payload))
	messageContainer.Refresh()
}

func GetTopicNames(db *db.Database) []string {
	var names []string
	for _, t := range db.Topics {
		names = append(names, t.TopicName)
	}
	return names
}
