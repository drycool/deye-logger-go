//go:build windows

package main

import "golang.org/x/sys/windows"

func init() {
	_ = windows.SetConsoleOutputCP(65001) // CP_UTF8
}
