# SWM

SWM is simple stacking window manager for X.

## How to use

For now, there is one `built-in` keyboard shotrcut, `alt + Return` which runs `xterm`.

SWM is meant to be controlled using X events and `swmctl`, its custom command sending tool.

X events could be sent using utilities like [xdotool](https://github.com/jordansissel/xdotool) or [wmctrl](http://tripie.sweb.cz/utils/wmctrl/). Their commands, as well as `swmctl` ones could be mapped to keyboard shortcuts ising utility like [sxhkd](https://github.com/baskerville/sxhkd). You can find example _sxhkd_ config in [examples](https://github.com/janbina/swm/tree/master/examples).

### swmctl

Swmctl is there for usecases which can't be done by sending X events. Those are:
- `shutdown` - shut down running instance of swm
- `move` - move active window (or window specified by id) by amount of pixels in specified direction (north/south/west/east) `swmctl move [-id XXX] [-n 20] [-s 20] [-w 20] [-e 20]`
- `resize` - enlarge/shrink active window (or window specified by id) by amount of pixels in specified direction (north/south/west/east) `swmctl resize [-id XXX] [-n 20] [-s 20] [-w 20] [-e 20]`
- `moveresize` - move and/or resize active window (or window specified by id) to specified location and to specified size. x/y coordinates as well is width/height can be specified relative to screen size, this could be used to tile widow left/right... `swmctl moveresize [-id XXX] [-g nw] [-x 0] [-y 0] [-wr .5] [-hr 1]` - tiles window left
- `set-desktop-names` - sets desktop names `swmctl set-desktop-names "a" "b" "c" "d"`
-	`move-drag-shortcut` - sets shortcut for moving windows using mouse
- `resize-drag-shortcut` - sets shortcut for resizing windows using mouse
- `begin-mouse-move` - calling this will initiate mouse move on window which is under the pointer
- `begin-mouse-resize` - calling this will initiate mouse resize on window which is under the pointer
- `cycle-win`, `cycle-win-rev` and `cycle-win-end` - used for cycling through windows on current desktop. `cycle-win` goes to second last active window, then third last and so on, `cycle-win-rev` goes in opposite direction. `cycle-win-end` needs to be called at the end of cycling operation to make temporary stacking/focus changes done while cycling permanent. Also, `cycle-win && cycle-win-end && cycle-win && cycle-win-end` results in something else than `cycle-win && cycle-win && cycle-win-end` - while the first ends up on window it started on, the second on third last screen of focus stack. `cycle-win-end` could be mapped to key release event of mod key used to cycle windows - commonly used `alt-tab` configuration could be seen in example sxhkd config
