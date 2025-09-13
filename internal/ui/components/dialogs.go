package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func FormDialog(title, confirmText, cancelText string, items []*widget.FormItem, onConfirm func(bool), parent fyne.Window) *dialog.FormDialog {
	d := dialog.NewForm(title, confirmText, cancelText, items, onConfirm, parent)
	d.Resize(fyne.NewSize(400, 200))
	return d
}

func ConfirmDialog(title, message string, onConfirm func(bool), parent fyne.Window) *dialog.ConfirmDialog {
	return dialog.NewConfirm(title, message, onConfirm, parent)
}

func ErrorDialog(err error, parent fyne.Window) {
	dialog.ShowError(err, parent)
}
