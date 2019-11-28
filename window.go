package main

func (h *harvester) renderMainWindow() error {
	return h.mainWindow.Load("http://" + h.listener.Addr().String() + "/templates/index.html")
}

func (h *harvester) sendErr(err error) {
	// errS := struct {
	// 	Type    string
	// 	Message string
	// }{
	// 	Type:    "error",
	// 	Message: err.Error(),
	// }

	h.mainWindow.Eval("console.log('here');")
}

func (h *harvester) sendTimers(timers TaskTimers) {
	// timerS := struct {
	// 	Type   string
	// 	Timers TaskTimers
	// }{
	// 	Type:   "renderTimers",
	// 	Timers: timers,
	// }

	// h.mainWindow.SendMessage(timerS)
}
