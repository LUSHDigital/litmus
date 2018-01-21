// Package p makes the color printers available throughout the application
package p

import "github.com/fatih/color"

var (
	Yellow = color.New(color.FgHiYellow).SprintFunc()
	Red    = color.New(color.FgHiRed).SprintFunc()
	Green  = color.New(color.FgHiGreen).SprintFunc()
	Blue   = color.New(color.FgHiBlue).SprintFunc()
)

