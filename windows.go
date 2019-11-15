package main

import (
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func (h *harvester) renderMainWindow() {
	h.mainWindow = h.app.NewWindow("Harvester")
	h.mainWindow.Resize(fyne.Size{Width: 400})
	h.mainWindow.SetPadded(false)
	h.mainWindow.SetOnClosed(func() {
		// @TODO: Add code to stop all timers
	})
	h.refresh()
	h.mainWindow.Show()
}

func (h *harvester) refresh() {
	h.mainWindow.SetContent(widget.NewVBox(
		widget.NewToolbar(
			widget.NewToolbarAction(
				theme.SettingsIcon(),
				func() {
					h.renderSettingsWindow()
				},
			),
			widget.NewToolbarSpacer(),
			widget.NewToolbarAction(
				theme.DocumentSaveIcon(),
				func() {
					h.app.Quit()
				},
			),
			widget.NewToolbarAction(
				theme.ConfirmIcon(),
				func() {
					h.app.Quit()
				},
			),
		),
		widget.NewLabel(time.Now().Format(time.Stamp)),
		widget.NewButton("Stop All", func() {
			// @TODO:  Add code to stop all started jiras
		}),
	))
}

func (h *harvester) renderSettingsWindow() {
	refreshInterval := widget.NewEntry()
	refreshInterval.SetText(h.settings.refreshInterval.String())
	jiraUser := widget.NewEntry()
	jiraUser.SetText(h.settings.jira.user)
	jiraPass := widget.NewPasswordEntry()
	jiraPass.SetText(h.settings.jira.pass)
	harvestUser := widget.NewEntry()
	harvestUser.SetText(h.settings.harvest.user)
	harvestPass := widget.NewPasswordEntry()
	harvestPass.SetText(h.settings.harvest.pass)

	errorMsg := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	errorBox := widget.NewHBox(
		widget.NewIcon(theme.WarningIcon()),
		errorMsg,
	)
	errorBox.Hide()

	h.settingsWindow = h.app.NewWindow("Settings")
	h.settingsWindow.Resize(fyne.Size{Width: 300})
	h.settingsWindow.SetFixedSize(true)
	h.settingsWindow.SetContent(
		widget.NewVBox(
			errorBox,
			widget.NewGroup(
				"General",
				widget.NewForm(
					&widget.FormItem{
						Text:   "Refresh Interval",
						Widget: refreshInterval,
					},
				),
			),
			widget.NewGroup(
				"Jira",
				widget.NewForm(
					&widget.FormItem{
						Text:   "Username",
						Widget: jiraUser,
					},
					&widget.FormItem{
						Text:   "Password",
						Widget: jiraPass,
					},
				),
			),
			widget.NewGroup(
				"Harvest",
				widget.NewForm(
					&widget.FormItem{
						Text:   "Username",
						Widget: harvestUser,
					},
					&widget.FormItem{
						Text:   "Password",
						Widget: harvestPass,
					},
				),
			),
			widget.NewButton("Submit", func() {
				h.settings.jira.user = jiraUser.Text
				h.settings.jira.pass = jiraPass.Text
				h.settings.harvest.user = harvestUser.Text
				h.settings.harvest.pass = harvestPass.Text

				interval, err := time.ParseDuration(refreshInterval.Text)
				if err != nil {
					errorBox.Show()
					errorMsg.SetText(err.Error())
					return
				}
				h.settings.refreshInterval = interval

				// Notify the parent process that something has changed
				h.changeCh <- true

				h.settingsWindow.Close()
			}),
		),
	)
	h.settingsWindow.Show()
}
