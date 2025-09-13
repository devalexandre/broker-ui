package main

import (
	"crypto/ed25519"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"github.com/devalexandre/broker-ui/internal/database"
	"github.com/devalexandre/broker-ui/internal/ui/views"
	"github.com/fynelabs/fyneselfupdate"
	"github.com/fynelabs/selfupdate"
)

func main() {
	// Initialize database
	db, err := database.New("./nats_servers.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create and run main window
	mainWindow := views.NewMainWindow(db)

	// Setup self-update (commented out for now to simplify testing)
	setupSelfUpdate(mainWindow.GetApp(), mainWindow.GetWindow())

	// Run the application
	mainWindow.Run()
}

func setupSelfUpdate(app fyne.App, window fyne.Window) {
	publicKey := ed25519.PublicKey{226, 162, 120, 210, 212, 122, 98, 250, 123, 180, 135, 69, 168, 77, 125, 41, 229, 245, 5, 32, 82, 254, 3, 37, 24, 224, 244, 63, 161, 123, 212, 197}

	httpSource := selfupdate.NewHTTPSource(nil, "https://geoffrey-artefacts.fynelabs.com/self-update/5d/5de43cb1-be73-4588-9c48-b2acb6169de0/{{.OS}}-{{.Arch}}/{{.Executable}}{{.Ext}}")

	config := fyneselfupdate.NewConfigWithTimeout(app, window, time.Minute, httpSource, selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * 12}, publicKey)

	_, err := selfupdate.Manage(config)
	if err != nil {
		fyne.LogError("Failed to set up update manager", err)
	}
}
