# SWM

[![Go Report Card](https://goreportcard.com/badge/github.com/janbina/swm)](https://goreportcard.com/report/github.com/janbina/swm)

SWM is simple stacking window manager for X.

## Installation

Clone this repository and run:
```
go build github.com/janbina/swm/cmd/swm
go build github.com/janbina/swm/cmd/swmctl
```

## How to use

For now, there is one built-in keyboard shotrcut, `ctrl + alt + Return`, which runs `xterm`.

SWM is meant to be controlled using X events and `swmctl`, its custom command sending tool.

X events could be sent using utilities like [xdotool](https://github.com/jordansissel/xdotool)
or [wmctrl](http://tripie.sweb.cz/utils/wmctrl/).
Their commands, as well as `swmctl` ones could be mapped to keyboard shortcuts using utility
like [sxhkd](https://github.com/baskerville/sxhkd).
You can find example _sxhkd_ config in [examples](https://github.com/janbina/swm/tree/master/examples).

### swmctl

Swmctl is there for usecases that can't be easily done sending X events.

##### General
- `swmctl shutdown` - shut down swm

##### Moving and resizing
- `swmctl move [-id windowID] [-n num] [-s num] [-w num] [-e num]`
    - move window by amount of pixels in specified direction (north/south/west/east)
    - windowId is optional and defaults to active (focused) window
- `swmctl resize [-id windowID] [-n num] [-s num] [-w num] [-e num]`
    - enlarge/shrink window by amount of pixels in specified direction (north/south/west/east)
    - windowId is optional and defaults to active (focused) window
- `swmctl moveresize [-id windowID] [-o origin]
                     [-x num] [-y num] [-xr num] [-yr num]
                     [-w num] [-h num] [-wr num] [-hr num]`
    - move and/or resize window to specified location and to specified size.
    - windowId is optional and defaults to active (focused) window
    - x/y coordinates as well is width/height could be specified relative to the screen size
    - x/y coordinates defaults to 0
    - if width/height is not specified, it won't be changed
    - origin specifies point on the screen to which x/y coordinates are related
        - possible values are `n, s, w, e, c` (north, south, west, east, center)
        - defaults to `nw` - top left corner
        - `-o se -x 10 -y 10` will place the window to bottom right corner with 10 pixels gap
        - `-o c -x 0 -y 0` will place the window to the center of the screen
    - this could be used to tile widow left/right
        - `swmctl moveresize -o nw -wr .5 -hr 1` tiles window left
- `swmctl begin-mouse-move`
    - initiate mouse move on window that is under the pointer
- `swmctl begin-mouse-resize`
    - initiate mouse resize on window that is under the pointer

##### Cycling windows
- Cycling through windows is done using swmctl commands `cycle-win`, `cycle-win-rev` and `cycle-win-end`.
- `cycle-win` goes to second last active window, then third last and so on, `cycle-win-rev` goes in opposite direction.
- `cycle-win-end` needs to be called at the end of cycling operation to make temporary stacking/focus changes done while cycling permanent.
- see example sxhkd config for commonly used `alt-tab` cycling shortcut configuration

##### Groups
Swm is using groups instead of more common virtual desktops model.
You can have arbitrary number of groups.
All or none groups could be visible at a time, with exception of _sticky_ group, that is always visible.
Window can be member of arbitrary number of groups (but at least one).
Window is visible if at least one of its groups is visible, otherwise hidden.

- `swmctl group mode (sticky|auto)`
    - set default group mode - when sticky, new windows are always assigned to sticky group, when auto, window preference is used with fallback of current group
- `swmctl group (toggle|show|hide|only) <groupId>`
    - change visibility of group - toggle it, show/hide it, or show it while hiding all others
- `swmctl group (set|add|remove) [-id windowId] [-g groupId]`
    - change group membership of window
    - windowId is optional and defaults to active (focused) window
    - groupId is optional and defaults to current group (group which is visible and was made visible most recently)
    - upon execution, info box is shown in top left corner of window listing all groups the window is member of
- `swmctl group names <name> [name...]`
    - set names for groups
    - example: `swmctl group names 1 2 3 4 5 6`
- `swmctl group get-visible`
    - get list of visible groups
    - returns group IDs separated by new-line and in ascending order
- `swmctl group get [-id windowId]`
    - get list of groups the window is member of
    - returns group IDs separated by new-line and in ascending order
    - windowId is optional and defaults to active (focused) window

##### Configuration
- `swmctl config border <width> <colorNormal> <colorActive> <colorUrgent>`
    - configure window border width and color for each of three states
    - set all borders at once or use `border-top`, `border-bottom`, `border-left`, `border-right` variants to set them separately (different width, color)
    - example: `swmctl config border 1 B0BEC5 00BCD4 F44336`
- `swmctl config info-bg-color <color>` and `swmctl config info-text-color <color>`
    - set info box background and text color (info box is used to show group membership)
    - example: `swmctl config info-bg-color 00BCD4`
- `swmctl config font <fontpath>`
    - set font used by swm (for now, only usage is in group indicator box)
    - example: `swmctl config font "/usr/share/fonts/TTF/JetBrainsMono-Bold.ttf"`
- `swmctl config move-drag-shortcut` and `swmctl config resize-drag-shortcut`
    - set shortcut for moving/resizing windows using mouse
    - example: `swmctl config move-drag-shortcut Mod1-1`

### swmrc

Swmrc is shell script that is executed by swm upon startup.
It is a good place to configure swm (border color etc.).
Example swmrc script could be found in [examples](https://github.com/janbina/swm/tree/master/examples).

Swm looks for swmrc script at that location: `$XDG_CONFIG_HOME/swm/swmrc`
