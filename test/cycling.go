package main

func testCycling() int {
	errorCnt := 0
	winNum := 5
	last := winNum - 1
	wins := createWindows(winNum)

	// few cycles of window cycling, ending at the same window we started
	for i := 0; i < winNum*3; i++ {
		index := (last - i) % winNum
		if index < 0 {
			index += winNum
		}
		assertActive(wins[index], &errorCnt)
		cycle(1, false)
	}
	assertActive(wins[last], &errorCnt)

	// few cycles of backwards window cycling, ending at the same window we started
	for i := 0; i < winNum*3; i++ {
		index := (last + i) % winNum
		assertActive(wins[index], &errorCnt)
		reverseCycle(1, false)
	}
	assertActive(wins[last], &errorCnt)

	// switching between top two windows
	assertActive(wins[last], &errorCnt)
	cycle(1, true)
	assertActive(wins[last-1], &errorCnt)
	cycle(1, true)
	assertActive(wins[last], &errorCnt)
	cycle(1, true)
	assertActive(wins[last-1], &errorCnt)
	cycle(1, true)
	assertActive(wins[last], &errorCnt)

	// switching three windows back, back to top window, and
	assertActive(wins[last], &errorCnt)
	cycle(3, true)
	assertActive(wins[last-3], &errorCnt)
	cycle(1, true)
	assertActive(wins[last], &errorCnt)
	cycle(1, true)
	assertActive(wins[last-3], &errorCnt)

	destroyWindows(wins)
	wins = createWindows(winNum)

	// reverse cycling
	assertActive(wins[last], &errorCnt)
	reverseCycle(1, true)
	assertActive(wins[0], &errorCnt)
	cycle(1, true)
	assertActive(wins[last], &errorCnt)
	cycle(1, true)
	assertActive(wins[0], &errorCnt)
	reverseCycle(2, true)
	assertActive(wins[2], &errorCnt)
	cycle(1, true)
	assertActive(wins[0], &errorCnt)

	destroyWindows(wins)

	return errorCnt
}

func cycle(times int, end bool) {
	for i := 0; i < times; i++ {
		swmctl("cycle-win")
	}
	if end {
		swmctl("cycle-win-end")
	}
}

func reverseCycle(times int, end bool) {
	for i := 0; i < times; i++ {
		swmctl("cycle-win-rev")
	}
	if end {
		swmctl("cycle-win-end")
	}
}
