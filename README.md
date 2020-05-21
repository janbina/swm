# SWM
[![Build and test](https://github.com/janbina/swm/workflows/Build%20and%20test/badge.svg)](https://github.com/janbina/swm/actions?query=workflow%3A%22Build+and+test%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/janbina/swm)](https://goreportcard.com/report/github.com/janbina/swm)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/janbina/swm)](https://golang.org/doc/devel/release.html)

SWM is a simple stacking window manager for X.

## Installation

Clone this repository and run:
```
go build github.com/janbina/swm/cmd/swm
go build github.com/janbina/swm/cmd/swmctl
```

## How to use

SWM is meant to be controlled using X events and `swmctl`, its custom command sending tool.

X events could be sent using utilities like [xdotool](https://github.com/jordansissel/xdotool)
or [wmctrl](http://tripie.sweb.cz/utils/wmctrl/).
Their commands, as well as `swmctl` ones could be mapped to keyboard shortcuts using utility
like [sxhkd](https://github.com/baskerville/sxhkd).
You can find example _sxhkd_ config in [examples](https://github.com/janbina/swm/tree/master/examples).

### swmctl

Swmctl is there for usecases that can't be easily done sending X events.

### swmrc

Swmrc is shell script that is executed by swm upon startup.
It is a good place to configure swm (border color etc.).
Example swmrc script could be found in [examples](https://github.com/janbina/swm/tree/master/examples).

### Documentation

See [documentation](https://github.com/janbina/swm/blob/master/doc/swm.adoc) for list of available `swmctl` commands,
where to put `swmrc`, examples, and more. 
