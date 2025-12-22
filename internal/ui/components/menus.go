package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/devalexandre/broker-ui/icons"
)

func MainMenu(onAddServer, onToggleTheme, onExit func(), isDarkTheme bool) *fyne.Container {
	addServerButton := widget.NewButtonWithIcon("Add Server", icons.AddServerIcon(), onAddServer)
	addServerButton.Importance = widget.HighImportance

	themeButton := widget.NewButtonWithIcon("", icons.ThemeToggleIcon(isDarkTheme), onToggleTheme)
	themeButton.Importance = widget.LowImportance

	exitButton := widget.NewButtonWithIcon("", icons.ExitIcon(), onExit)
	exitButton.Importance = widget.LowImportance

	return container.NewPadded(container.NewHBox(addServerButton, layout.NewSpacer(), themeButton, exitButton))
}

func ServerMenu(onAddPublisher, onAddSubscriber, onDisconnect func()) *fyne.Container {
	addPublisher := widget.NewButtonWithIcon("Publisher", icons.PublisherIcon(), onAddPublisher)
	addSubscriber := widget.NewButtonWithIcon("Subscriber", icons.SubscriberIcon(), onAddSubscriber)
	disconnect := widget.NewButtonWithIcon("Disconnect", icons.ExitIcon(), onDisconnect)
	disconnect.Importance = widget.DangerImportance

	return container.NewHBox(addPublisher, addSubscriber, layout.NewSpacer(), disconnect)
}
