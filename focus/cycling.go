package focus


var cyclingState []FocusableWindow
var cyclableWindows []FocusableWindow

func CyclingFocus(state int) {
	if cyclingState == nil {
		cyclingState = make([]FocusableWindow, len(windows))
		copy(cyclingState, windows)
		cyclableWindows = make([]FocusableWindow, 0, len(windows))
		for _, win := range windows {
			if win.IsFocusable() {
				cyclableWindows = append(cyclableWindows, win)
			}
		}
	}

	if len(cyclableWindows) < 2 { // nothing to cycle
		return
	}

	index := (len(cyclableWindows) - 1 + (state % len(cyclableWindows))) % len(cyclableWindows)
	win := cyclableWindows[index]

	copy(windows, cyclingState)
	Focus(win)
}

func CyclingEnded() {
	cyclingState = nil
	cyclableWindows = nil
}
