//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// MEMORYSTATUSEX est la structure passée à GlobalMemoryStatusEx.
// https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-memorystatusex
type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
)

// availableMemoryBytes retourne la mémoire physique disponible sur Windows.
func availableMemoryBytes() uint64 {
	var ms memoryStatusEx
	ms.dwLength = uint32(unsafe.Sizeof(ms))
	globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	return ms.ullAvailPhys
}
