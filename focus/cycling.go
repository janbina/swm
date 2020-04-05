package focus

var cyclingState []FocusableWindow
var cyclableWindows []FocusableWindow

func CyclingFocus(state int) FocusableWindow {
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
		return nil
	}

	index := (len(cyclableWindows) - 1 + (state % len(cyclableWindows))) % len(cyclableWindows)
	win := cyclableWindows[index]

	copy(windows, cyclingState)
	Focus(win)
	return win
}

func CyclingEnded() FocusableWindow {
	if len(windows) == 0 {
		return nil
	}
	cyclingState = nil
	cyclableWindows = nil
	return windows[len(windows)-1]
}
