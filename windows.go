package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func (h *harvester) renderMainWindow() {
	h.mainWindow = h.app.NewWindow("Harvester")
	h.mainWindow.Resize(fyne.Size{Width: 400, Height: 100})
	h.mainWindow.SetPadded(false)
	h.refresh(false)
	h.mainWindow.Show()
}

func (h *harvester) redraw() {
	mainContent := widget.NewVBox(h.drawJiraObjects()...)
	if h.jiraMsg != "" {
		mainContent = widget.NewVBox(widget.NewLabelWithStyle(
			breakString(h.jiraMsg, 50),
			fyne.TextAlignCenter,
			fyne.TextStyle{
				Italic: true,
			},
		))
		h.jiraMsg = ""
	}

	h.mainWindow.SetContent(widget.NewVBox(
		widget.NewToolbar(
			widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
				h.getJiraListClipboard()
			}),
			widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
				h.refresh(false)
			}),
			widget.NewToolbarSpacer(),
			widget.NewToolbarAction(theme.InfoIcon(), func() {
				h.renderAboutWindow()
			}),
			widget.NewToolbarAction(theme.SettingsIcon(), func() {
				h.renderSettingsWindow()
			}),
		),
		mainContent,
		widget.NewHBox(
			widget.NewButtonWithIcon("Show Times", theme.InfoIcon(), func() {
				h.showJiraTimes(nil)
			}),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Stop All", theme.CancelIcon(), func() {
				for _, jira := range h.activeJiras {
					h.saveJiraTime(jira, "stop")
				}
				h.redraw()
			}),
		),
	))
}

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
			widget.NewButton("Submit", func() {
				h.settings.Jira.User = jiraUser.Text
				h.settings.Jira.Pass = jiraPass.Text
				h.settings.Jira.URL = strings.TrimSuffix(jiraURL.Text, "/")

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

func (h *harvester) getJiraListClipboard() {
	var clipboard string
	for _, jira := range h.activeJiras {
		clipboard += fmt.Sprintf("%s: %s\n", jira.Key, jira.Fields.Summary)
	}
	h.mainWindow.Clipboard().SetContent(clipboard)
}

func (h *harvester) renderAboutWindow() {
	h.aboutWindow = h.app.NewWindow("About")
	h.aboutWindow.SetContent(widget.NewHBox(
		widget.NewLabel("Version"),
		widget.NewLabel(version),
	))
	h.aboutWindow.Show()
}

func breakString(msg string, size int) string {
	for i := size; i < len(msg); i += (size + 1) {
		msg = msg[:i] + "\n" + msg[i:]
	}
	return msg
}
