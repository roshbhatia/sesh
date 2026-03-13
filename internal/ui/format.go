package ui

import "fmt"

// Success returns a green ✓ message.
func Success(msg string) string {
	return StyleSuccess.Render("✓") + " " + msg
}

// Error returns a red ✗ message.
func Error(msg string) string {
	return StyleError.Render("✗") + " " + msg
}

// Warning returns a yellow ⚠ message.
func Warning(msg string) string {
	return StyleWarning.Render("⚠") + " " + msg
}

// Info returns a blue ℹ message.
func Info(msg string) string {
	return StyleInfo.Render("ℹ") + " " + msg
}

// Faint dims the given string.
func Faint(msg string) string {
	return Dim.Render(msg)
}

// Successf is a formatted version of Success.
func Successf(format string, a ...any) string {
	return Success(fmt.Sprintf(format, a...))
}

// Errorf is a formatted version of Error.
func Errorf(format string, a ...any) string {
	return Error(fmt.Sprintf(format, a...))
}

// Warningf is a formatted version of Warning.
func Warningf(format string, a ...any) string {
	return Warning(fmt.Sprintf(format, a...))
}

// Infof is a formatted version of Info.
func Infof(format string, a ...any) string {
	return Info(fmt.Sprintf(format, a...))
}
