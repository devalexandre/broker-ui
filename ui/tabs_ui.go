package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2/container"
	"github.com/devalexandre/brokers-ui/db"
)

// AddTabsForTopicsAndSubs adiciona abas para os tópicos e subscrições carregados
func AddTabsForTopicsAndSubs(database *db.Database) {
	log.Printf("Adicionando abas para tópicos e subscrições")

	log.Printf("Adicionando abas para %v tópicos e %v subscrições", len(database.Topics), len(database.Subs))
	for _, t := range database.Topics {
		topicName := fmt.Sprintf("topic-%v", t.TopicName)
		tab := container.NewTabItem(topicName, CreateTopicTabContent(t.TopicName))
		TabContainer.Append(tab)
	}

	for _, s := range database.Subs {
		subName := fmt.Sprintf("sub-%v", s.SubName)
		tab := container.NewTabItem(subName, CreateSubTabContent(s.SubName))
		TabContainer.Append(tab)
	}

	TabContainer.Refresh()
}
