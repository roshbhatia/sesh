package ui

import "fmt"

// Success returns a green-styled success message with ✓ prefix.
func Success(msg string) string {
	return DefaultStyles.Success.Render("✓ " + msg)
}

// Error returns a red-styled error message with ✗ prefix.
func Error(msg string) string {
	return DefaultStyles.Error.Render("✗ " + msg)
}

// Warning returns a yellow-styled warning message with ⚠ prefix.
func Warning(msg string) string {
	return DefaultStyles.Warning.Render("⚠ " + msg)
}

// Info returns a blue-styled informational message.
func Info(msg string) string {
	return DefaultStyles.Info.Render(msg)
}

// Dim returns a dimmed/muted string.
func Dim(msg string) string {
	return DefaultStyles.Dim.Render(msg)
}

// Bold returns a bold string.
func Bold(msg string) string {
	return DefaultStyles.Bold.Render(msg)
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
