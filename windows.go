package main

import (
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func (h *harvester) renderMainWindow() {
	h.mainWindow = h.app.NewWindow("Harvester")
	h.mainWindow.Resize(fyne.Size{Width: 400, Height: 100})
	h.mainWindow.SetPadded(false)
	h.mainWindow.SetOnClosed(func() {
		// @TODO: Add code to stop all timers
	})
	h.refresh()
	h.mainWindow.Show()
}

func (h *harvester) redraw() {
	h.mainWindow.SetContent(widget.NewVBox(
		widget.NewToolbar(
			widget.NewToolbarAction(
				theme.DocumentSaveIcon(),
				func() {
					h.showJiraTimes()
				},
			),
			widget.NewToolbarAction(
				theme.ViewRefreshIcon(),
				func() {
					h.refresh()
				},
			),
			widget.NewToolbarSpacer(),
			widget.NewToolbarAction(
				theme.SettingsIcon(),
				func() {
					h.renderSettingsWindow()
				},
			),
		),
		widget.NewVBox(h.drawJiraObjects()...),
		widget.NewButton("Stop All", func() {
			for _, jira := range h.activeJiras {
				h.saveJiraTime(jira.Key, "stop")
			}
			h.refresh()
		}),
	))
}

func (h *harvester) renderSettingsWindow() {
	refreshInterval := widget.NewEntry()
	refreshInterval.SetText(h.settings.RefreshInterval.String())

	themeSelector := widget.NewCheck("Dark Mode", func(checked bool) {
		if checked {
			h.app.Settings().SetTheme(theme.DarkTheme())
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
	harvestUser := widget.NewEntry()
	harvestUser.SetText(h.settings.Harvest.User)
	harvestPass := widget.NewPasswordEntry()
	harvestPass.SetText(h.settings.Harvest.Pass)

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
				h.settings.Jira.User = jiraUser.Text
				h.settings.Jira.Pass = jiraPass.Text
				h.settings.Jira.URL = strings.TrimSuffix(jiraURL.Text, "/")
				h.settings.Harvest.User = harvestUser.Text
				h.settings.Harvest.Pass = harvestPass.Text

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

				h.settingsWindow.Close()
			}),
		),
	)
	h.settingsWindow.Show()
}
