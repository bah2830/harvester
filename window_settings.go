package main

import (
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func (h *harvester) renderSettingsWindow() {
	refreshInterval := widget.NewEntry()
	refreshInterval.SetText(h.settings.RefreshInterval.String())

	themeSelector := widget.NewCheck("Dark Mode", func(checked bool) {
		if checked {
			h.app.Settings().SetTheme(defaultTheme())
		} else {
			h.app.Settings().SetTheme(theme.LightTheme())
		}
	})
	if h.settings.DarkTheme {
		themeSelector.SetChecked(true)
	} else {
		themeSelector.SetChecked(false)
	}

	jiraURL := widget.NewEntry()
	jiraURL.SetText(h.settings.Jira.URL)
	jiraUser := widget.NewEntry()
	jiraUser.SetText(h.settings.Jira.User)
	jiraPass := widget.NewPasswordEntry()
	jiraPass.SetText(h.settings.Jira.Pass)

	harvestAccount := widget.NewEntry()
	harvestAccount.SetText(h.settings.Harvest.User)
	harvestToken := widget.NewPasswordEntry()
	harvestToken.SetText(h.settings.Harvest.Pass)

	errorMsg := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	errorBox := widget.NewHBox(
		widget.NewIcon(theme.WarningIcon()),
		errorMsg,
	)
	errorBox.Hide()

	w := h.app.NewWindow("Settings")
	w.Resize(fyne.Size{Width: 300})
	w.SetFixedSize(true)
	w.SetContent(
		widget.NewVBox(
			errorBox,
			widget.NewGroup(
				"General",
				themeSelector,
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
						Text:   "URL",
						Widget: jiraURL,
					},
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
						Text:   "Account ID",
						Widget: harvestAccount,
					},
					&widget.FormItem{
						Text:   "Token",
						Widget: harvestToken,
					},
				),
			),
			widget.NewButton("Submit", func() {
				h.settings.Jira.User = jiraUser.Text
				h.settings.Jira.Pass = jiraPass.Text
				h.settings.Jira.URL = strings.TrimSuffix(jiraURL.Text, "/")

				h.settings.Harvest.User = harvestAccount.Text
				h.settings.Harvest.Pass = harvestToken.Text

				interval, err := time.ParseDuration(refreshInterval.Text)
				if err != nil {
					errorBox.Show()
					errorMsg.SetText(err.Error())
					return
				}
				h.settings.RefreshInterval = interval
				h.settings.DarkTheme = themeSelector.Checked

				// Notify the parent process that something has changed
				h.changeCh <- true

				w.Close()
			}),
		),
	)
	w.Show()
}
