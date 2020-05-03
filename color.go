package gomake

import (
	"fmt"
	"github.com/mgutz/ansi"
)

type Color func(string) string

var HGreen Color = ansi.ColorFunc("green+h")
var HYellow Color = ansi.ColorFunc("yellow+h")
var HCyan Color = ansi.ColorFunc("cyan+h")
var HRed Color = ansi.ColorFunc("red+h")
var HMagenta Color = ansi.ColorFunc("magenta+h")
var HBlue Color = ansi.ColorFunc("blue+h")

var Green Color = ansi.ColorFunc("green")
var Yellow Color = ansi.ColorFunc("yellow")
var Cyan Color = ansi.ColorFunc("cyan")
var Red Color = ansi.ColorFunc("red")
var Magenta Color = ansi.ColorFunc("magenta")
var Blue Color = ansi.ColorFunc("blue")

func ColorPrintln(text string, color Color) {
	fmt.Println(color(text))
}
