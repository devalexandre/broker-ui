package components

import (
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewSuccessButton cria um botão com o estilo de "sucesso" (verde)
func NewSuccessButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.HighImportance // Define a importância como "Alta"
	btn.SetIcon(theme.ConfirmIcon())       // Opcional: define um ícone adequado
	return btn
}

// NewInfoButton cria um botão com o estilo de "informação" (azul)
func NewInfoButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.HighImportance // Define a importância como "Alta"
	btn.SetIcon(theme.InfoIcon())          // Opcional: define um ícone adequado
	return btn
}

// NewWarningButton cria um botão com o estilo de "aviso" (laranja)
func NewWarningButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.HighImportance // Define a importância como "Alta"
	btn.SetIcon(theme.WarningIcon())       // Opcional: define um ícone adequado
	return btn
}

// NewDangerButton cria um botão com o estilo de "perigo" (vermelho)
func NewDangerButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.DangerImportance // Define a importância como "Perigo"
	btn.SetIcon(theme.CancelIcon())          // Opcional: define um ícone adequado

	return btn
}
