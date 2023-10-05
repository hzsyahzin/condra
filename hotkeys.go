// For example usage check test file

package main

import windows "golang.org/x/sys/windows"

// Struct containing key name and virtual keycode
type Keycode struct {
	Key    string
	Code   uintptr
}

// Gets hotkey detecting procedure from user32.dll
func GetAsyncKeyStateProc() *windows.Proc {
	user32 := windows.MustLoadDLL("user32")
	defer user32.Release()

	getAsyncKeyState := user32.MustFindProc("GetAsyncKeyState")
	return getAsyncKeyState
}