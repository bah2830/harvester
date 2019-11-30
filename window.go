package main

import (
	"encoding/json"
	"log"
	"strings"
	"text/template"

	"github.com/bah2830/webview"
	"github.com/skratchdot/open-golang/open"
)

func (h *harvester) renderMainWindow() {
	h.mainWindow = webview.New(webview.Settings{
		Title:                  "Harvester",
		Resizable:              true,
		Height:                 400,
		Width:                  400,
		URL:                    "http://" + h.listener.Addr().String() + "/templates/main.html",
		ExternalInvokeCallback: h.handleMainRPC,
		Debug:                  h.debug,
	})

	h.mainWindow.Dispatch(func() {
		h.injectDefaults(h.mainWindow)

		h.mainWindow.Bind("timers", h.Timers)
		h.mainWindow.Eval(string(MustAsset("resources/js/main.js")))
		h.mainWindow.Eval("init()")
	})
}

func (h *harvester) renderSettings() {
	if h.settingsWindow != nil {
		return
	}

	h.settingsWindow = webview.New(webview.Settings{
		Title:                  "Settings",
		Resizable:              false,
		Height:                 500,
		Width:                  400,
		URL:                    "http://" + h.listener.Addr().String() + "/templates/settings.html",
		ExternalInvokeCallback: h.handleSettingsRPC,
		Debug:                  h.debug,
	})

	h.settingsWindow.Dispatch(func() {
		h.injectDefaults(h.settingsWindow)

		h.settingsWindow.Bind("settings", h.Settings)
		h.settingsWindow.Eval(string(MustAsset("resources/js/settings.js")))
		h.settingsWindow.Eval("init()")
	})

	h.settingsWindow.Run()
	h.settingsWindow = nil
}

func (h *harvester) renderInfo() {
	if h.infoWindow != nil {
		return
	}

	timesHTML, err := h.getTimeSheetHTML()
	if err != nil {
		h.sendErr(err)
		return
	}

	h.infoWindow = webview.New(webview.Settings{
		Title:     "Info",
		Resizable: false,
		Height:    500,
		Width:     600,
		URL:       "http://" + h.listener.Addr().String() + "/templates/info.html",
		Debug:     h.debug,
	})

	h.infoWindow.Dispatch(func() {
		h.injectDefaults(h.infoWindow)
		h.infoWindow.Eval("$('#info-container').html('" + template.JSEscapeString(timesHTML) + "')")
	})

	h.infoWindow.Run()
	h.infoWindow = nil
}

func (h *harvester) injectDefaults(w webview.WebView) {
	w.InjectCSS(string(MustAsset("resources/css/bootstrap/bootstrap.min.css")))
	w.InjectCSS(string(MustAsset("resources/css/main.css")))
	w.Eval(string(MustAsset("resources/js/jquery/jquery-3.4.1.min.js")))
	w.Eval(string(MustAsset("resources/js/bootstrap/bootstrap.bundle.min.js")))
}

func (h *harvester) handleMainRPC(w webview.WebView, data string) {
	switch {
	case data == "copy":
	case data == "refresh":
		h.Refresh()
	case data == "info":
		h.renderInfo()
	case data == "settings":
		h.renderSettings()
	case strings.Contains(data, "|"):
		parts := strings.Split(data, "|")
		task, err := h.Timers.Tasks.GetByKey(parts[1])
		if err != nil {
			h.sendErr(err)
			return
		}

		var currentRunning string
		switch parts[0] {
		case "start":
			currentRunning = task.Key
			err = task.Start(h.db, h.harvestClient)
		case "stop":
			err = task.Stop(h.db, h.harvestClient)
		case "open":
			err = open.Run(h.Settings.Jira.URL + "/browse/" + parts[1])
		}
		if err != nil {
			h.sendErr(err)
			return
		}

		h.updateTimers(currentRunning)
		h.sendTimers()
	default:
		log.Println("unknown rpc handler " + data)
	}
}

func (h *harvester) handleSettingsRPC(w webview.WebView, data string) {
	var settings Settings
	if err := json.Unmarshal([]byte(data), &settings); err != nil {
		h.sendErr(err)
		return
	}

	h.Settings.Jira.URL = settings.Jira.URL
	h.Settings.Jira.User = settings.Jira.User
	h.Settings.Jira.Pass = settings.Jira.Pass
	h.Settings.Harvest.User = settings.Harvest.User
	h.Settings.Harvest.Pass = settings.Harvest.Pass
	h.changeCh <- true
}

func (h *harvester) sendErr(err error) {
	h.mainWindow.Dialog(webview.DialogTypeAlert, webview.DialogFlagError, "Error", err.Error())
}

func (h *harvester) sendTimers() {
	h.mainWindow.Dispatch(func() {
		h.mainWindow.Bind("timers", h.Timers)
		h.mainWindow.Eval("timerUpdate()")
	})
}
