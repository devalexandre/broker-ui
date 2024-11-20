package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/devalexandre/brokers-ui/db"
	"log"

	"fyne.io/fyne/v2/container"
)

// AddTabsForTopicsAndSubs adiciona abas para os tópicos e subscrições carregados
func AddTabsForTopicsAndSubs(window fyne.Window, db *db.Database, serverID int) {
	log.Printf("Adicionando abas para tópicos e subscrições")

	log.Printf("Adicionando abas para %v tópicos e %v subscrições", len(*Topics), len(*Subs))
	for _, t := range *Topics {
		topicName := fmt.Sprintf("topic-%v", t.TopicName)
		tab := container.NewTabItem(topicName, CreateTopicTabContent(window, t.TopicName, db, serverID))
		TabContainer.Append(tab)
	}

	for _, s := range *Subs {
		subName := fmt.Sprintf("sub-%v", s.SubName)
		tab := container.NewTabItem(subName, CreateSubTabContent(s.SubName))
		TabContainer.Append(tab)
	}

	TabContainer.Refresh()
}
