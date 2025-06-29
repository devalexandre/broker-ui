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
	"github.com/devalexandre/broker-ui/natscli"
	"github.com/devalexandre/broker-ui/themes/dracula"
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
	ID       int
	ServerID int
	SubName  string
}

var servers []server
var topics []topic
var subs []sub
var natsServers = make(map[int]*natscli.Nats)

// Um mapa para armazenar as mensagens enviadas para cada tópico e sub
var sentMessages = make(map[string][]string)
var receivedMessages = make(map[string][]string)
var dasboardReceivedMessages = make(map[string]int)

var myWindow fyne.Window

func main() {

	// Inicializar o banco de dados SQLite3
	var err error
	db, err = sql.Open("sqlite3", "./nats_servers.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Criar tabelas
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, url TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS topics (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, topic_name TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS subs (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, sub_name TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	// Carregar servidores do banco de dados
	loadServers()

	myApp := app.New()
	myApp.Settings().SetTheme(dracula.DraculaTheme{})
	myWindow = myApp.NewWindow("NATS Client")

	// Menu Superior
	menu := container.NewBorder(
		nil, nil,
		widget.NewButtonWithIcon("Add Server", theme.ContentAddIcon(), func() {
			addServer(myWindow)
		}),
		widget.NewButtonWithIcon("Exit", theme.CancelIcon(), func() {
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

	// Painel Principal
	tabContainer = container.NewAppTabs()
	welcomeTab := container.NewTabItem("Welcome", welcomeMessage)
	tabContainer.Append(welcomeTab)
	// Lista de Servidores
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

	// Layout Principal
	mainContent := container.NewHSplit(
		container.NewVBox(serverList),
		tabContainer,
	)
	mainContent.Offset = 0.2 // Define o espaço do menu esquerdo

	content := container.NewBorder(menu, nil, nil, nil, mainContent)

	//update
	selfManage(myApp, myWindow)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()

}

func displayServerOptions(window fyne.Window, name string, url string) {
	// Limpar as abas antes de recriar
	clearTabs()

	// Menu adicional para o servidor selecionado
	menu := container.NewHBox(
		widget.NewButtonWithIcon("Add Topic", theme.ContentAddIcon(), func() {
			addTopic(window, selectedServerID)
		}),
		widget.NewButtonWithIcon("Add Sub", theme.ContentAddIcon(), func() {
			addSub(window, selectedServerID)
		}),
		widget.NewButtonWithIcon("Disconnect", theme.MediaStopIcon(), func() {
			disconnectFromServer()
		}),
	)

	// Botão de Editar Conexão
	editButton := widget.NewButtonWithIcon("Edit Connection", theme.ViewRefreshIcon(), func() {
		editServerConnection(window, selectedServerID, name, url)
	})

	panel := container.NewVBox(
		menu,
		widget.NewLabel(fmt.Sprintf("Connected to %s (%s)", name, url)),
		editButton,
	)

	// Tentativa de criar um novo cliente NATS
	natsServers[selectedServerID], NatsError = natscli.NewNats(url)
	if NatsError != nil {
		dialog.ShowError(NatsError, window)
	} else {
		// Carregar tópicos e subs se a conexão for bem-sucedida
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
	// Fecha a conexão NATS e limpa as tabs
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
				loadServers()        // Recarrega a lista de servidores
				serverList.Refresh() // Atualiza a lista exibida
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
		tab := container.NewTabItemWithIcon(subname, theme.ViewRefreshIcon(), createSubTabContent(s.SubName))
		tabContainer.Append(tab)
	}

	tabContainer.Refresh()
}

func createSubTabContent(subName string) fyne.CanvasObject {
	// Canal para receber mensagens
	messageChan := make(chan string)

	// Caixa vertical para armazenar as mensagens
	messageContainer := container.NewVBox()

	receivedMessages[subName] = []string{} // Inicializar lista de mensagens recebidas para o sub

	// Goroutine para lidar com a assinatura NATS e enviar mensagens para o canal
	go func() {
		err := natsServers[selectedServerID].Subscribe(subName, func(m *nats.Msg) {
			payload := string(m.Data)
			log.Printf("Received message from sub %s: %s", subName, payload)

			// Enviar a mensagem para o canal
			messageChan <- payload
		})

		if err != nil {
			log.Printf("Error subscribing to subject %s: %v", subName, err)
			return
		}
	}()

	// Goroutine para monitorar o canal e atualizar a interface gráfica
	go func() {
		for payload := range messageChan {
			receivedMessages[subName] = append(receivedMessages[subName], payload)

			// Adicionar a mensagem ao container na thread principal
			messageContainer.Add(widget.NewLabel(payload))

			// Atualizar a interface gráfica
			messageContainer.Refresh()
		}
	}()

	return container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Sub: %s", subName)),
		messageContainer,
	)
}

func createTopicTabContent(topicName string, serverId int) fyne.CanvasObject {
	// Slice para armazenar as mensagens
	var messages []string

	// Container vertical para exibir as mensagens
	messageContainer := container.NewVBox()

	// Entrada de texto para a mensagem
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Enter message payload here...")

	// Botão de envio
	sendButton := widget.NewButton("Send", func() {
		payload := messageEntry.Text
		sendMessageToTopic(topicName, payload, messageContainer, &messages)

		// Limpar a entrada de mensagem
		messageEntry.SetText("")
	})

	// Cria o botão "X" para fechar a aba
	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		dialog.NewConfirm("Delete Topic", "Are you sure you want to delete this topic?", func(confirmed bool) {
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
			widget.NewLabel(fmt.Sprintf("Topic: %s", topicName)),
			closeButton,
		),
		messageEntry,
		sendButton,
		messageContainer,
	)
}

func sendMessageToTopic(topicName string, payload string, messageContainer *fyne.Container, messages *[]string) {
	if payload == "" {
		return
	}

	// Enviar a mensagem ao NATS server
	err := natsServers[selectedServerID].Publish(topicName, []byte(payload))
	if err != nil {
		fmt.Println("Error publishing message:", err)
		return
	}

	log.Printf("Sending message to topic %s: %s", topicName, payload)

	// Armazena a mensagem na lista para exibição
	*messages = append(*messages, payload)

	// Atualizar a interface gráfica
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

	// Inserir Nome e URL no banco de dados SQLite3
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
	// Carrega os servidores do banco de dados
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
				saveSub(serverID, entry.Text)
				displayServerOptions(myWindow, selectedServer.Name, selectedServer.URL)
			}
		},
		window,
	)
	dialog.Resize(fyne.NewSize(400, 200))
	dialog.Show()
}

func saveSub(serverID int, subName string) {
	if subName == "" {
		return
	}

	stmt, err := db.Prepare("INSERT INTO subs(server_id, sub_name) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID, subName)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Sub saved:", subName)
}

func loadSubs(serverID int) {
	rows, err := db.Query("SELECT id, sub_name FROM subs WHERE server_id = ?", serverID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	subs = []sub{}
	for rows.Next() {
		var s sub
		err := rows.Scan(&s.ID, &s.SubName)
		if err != nil {
			log.Fatal(err)
		}
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

// Adicionar a aba de Dashboard
func addDashboardTab() {
	// Variáveis para armazenar contagens de mensagens por sub
	metrics := make(map[string]*widget.Label)

	// Cria um container para os gráficos ou tabelas de cada sub
	dashboardContainer := container.NewVBox(
		widget.NewLabel("Message Monitoring Dashboard"),
	)

	// Adiciona uma seção para cada sub com um contador de mensagens
	for _, sub := range subs {
		label := widget.NewLabel(fmt.Sprintf("Sub: %s - Messages received: 0", sub.SubName))
		metrics[sub.SubName] = label
		dashboardContainer.Add(label)
	}

	// Adiciona a aba do Dashboard ao tabContainer
	dashboardTab := container.NewTabItem("Dashboard", dashboardContainer)
	tabContainer.Append(dashboardTab)
	tabContainer.Select(dashboardTab)

	// Goroutine para monitorar as mensagens e atualizar as métricas
	go func() {
		for {
			for subName, label := range metrics {
				// Conta o número de mensagens recebidas para esta sub
				count := len(receivedMessages[subName])
				label.SetText(fmt.Sprintf("Sub: %s - Messages received: %d", subName, count))
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func selfManage(a fyne.App, w fyne.Window) {
	publicKey := ed25519.PublicKey{200, 37, 164, 131, 164, 172, 52, 181, 239, 251, 200, 30, 190, 92, 215, 209, 174, 6, 144, 222, 75, 0, 52, 75, 52, 11, 58, 59, 217, 7, 46, 75}

	// The public key above matches the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "https://geoffrey-artefacts.fynelabs.com/self-update/51/510d1864-0874-460d-bce6-36438c777ed4/{{.OS}}-{{.Arch}}/{{.Executable}}{{.Ext}}")

	config := fyneselfupdate.NewConfigWithTimeout(a, w, time.Minute, httpSource, selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * 12}, publicKey)

	_, err := selfupdate.Manage(config)
	if err != nil {
		fyne.LogError("Failed to set up update manager", err)
	}
}
