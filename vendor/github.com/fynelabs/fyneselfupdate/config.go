package fyneselfupdate

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/fynelabs/selfupdate"
)

// NewConfig create a selfupdate.Config structure that provide Fyne UI dialog to progress the update
func NewConfig(app fyne.App, win fyne.Window, source selfupdate.Source, schedule selfupdate.Schedule, publicKey ed25519.PublicKey) *selfupdate.Config {
	return NewConfigWithTimeout(app, win, 0, source, schedule, publicKey)
}

// NewConfigWithTimeout create a selfupdate.Config structure that provide Fyne UI dialog to progress the update
// and provide a "default to yes" behavior if the user doesn't interact within timeout time to provide unattended
// update.
func NewConfigWithTimeout(app fyne.App, win fyne.Window, timeout time.Duration, source selfupdate.Source, schedule selfupdate.Schedule, publicKey ed25519.PublicKey) *selfupdate.Config {
	selfupdate.LogError = func(format string, args ...interface{}) {
		err := errors.New(fmt.Sprintf(format, args))
		fyne.Do(func() {
			dialog.ShowError(err, win)
		})
	}
	return &selfupdate.Config{
		Source:    source,
		Schedule:  schedule,
		PublicKey: publicKey,

		UpgradeConfirmCallback: NewUpgradeConfirmCallbackWithTimeout(win, timeout),
		ProgressCallback:       NewProgressCallback(win),
		RestartConfirmCallback: NewRestartConfirmCallbackWithTimeout(win, timeout),
		ExitCallback:           NewExitCallback(app, win),
	}
}
