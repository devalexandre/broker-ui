package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/brokers-ui/components"
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
				RefreshTopicsAndSubs(serverID, db)
				AddTabsForTopicsAndSubs(window, db, serverID)

				fmt.Println(fmt.Sprintf("Topic %s added", entry.Text))
				fmt.Println(fmt.Sprintf("Loaded %d topics for server %d", len(*Topics), serverID))

			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func CreateTopicTabContent(window fyne.Window, topicName string, db *db.Database, serverID int) fyne.CanvasObject {
	var messages []string

	messageContainer := container.NewVBox()

	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Enter message payload here...")

	sendButton := widget.NewButton("Send", func() {
		payload := messageEntry.Text
		SendMessageToTopic(topicName, payload, messageContainer, &messages)

		messageEntry.SetText("")
	})

	dialogDelete := dialog.NewConfirm("Delete Topic", "Are you sure you want to delete this topic?", func(confirmed bool) {
		if confirmed {
			err := db.DeleteTopic(serverID, topicName)

			if err != nil {
				log.Printf("Error deleting topic: %v", err)
				return
			}

			RefreshTopicsAndSubs(SelectedServerID, db)
			//AddTabsForTopicsAndSubs(window, db, serverID)
		}
	}, window)

	deleteButton := components.NewDangerButton("Delete", func() {
		dialogDelete.Show()
	})

	hbuttons := container.NewHBox(
		sendButton,
		layout.NewSpacer(), // Adiciona espaçamento entre os botões
		deleteButton,
	)

	return container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Topic: %s", topicName)),
		messageEntry,
		hbuttons,
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

func GetTopicNames() []string {
	var names []string
	for _, t := range *Topics {
		names = append(names, t.TopicName)
	}
	return names
}
