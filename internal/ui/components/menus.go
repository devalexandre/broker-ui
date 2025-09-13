package components

import (
"fyne.io/fyne/v2"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/widget"
"github.com/devalexandre/broker-ui/icons"
)

func MainMenu(onAddServer, onToggleTheme, onExit func(), isDarkTheme bool) *fyne.Container {
	addServerButton := widget.NewButtonWithIcon("Add Server", icons.AddServerIcon(), onAddServer)
	themeButton := widget.NewButtonWithIcon("Theme", icons.ThemeToggleIcon(isDarkTheme), onToggleTheme)
	exitButton := widget.NewButtonWithIcon("Exit", icons.ExitIcon(), onExit)

	return container.NewBorder(
nil, nil,
container.NewHBox(addServerButton, themeButton),
exitButton,
)
}

func ServerMenu(onAddPublisher, onAddSubscriber, onDisconnect func()) *fyne.Container {
	return container.NewHBox(
widget.NewButtonWithIcon("Add Publisher", icons.PublisherIcon(), onAddPublisher),
		widget.NewButtonWithIcon("Add Subscriber", icons.SubscriberIcon(), onAddSubscriber),
		widget.NewButtonWithIcon("Disconnect", icons.ExitIcon(), onDisconnect),
	)
}
