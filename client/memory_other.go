//go:build !windows

package main

import "runtime"

// availableMemoryBytes retourne une estimation de la mémoire disponible
// sur les systèmes non-Windows, en utilisant runtime.MemStats.
func availableMemoryBytes() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	// Sys : mémoire totale obtenue de l'OS ; HeapInuse : mémoire tas utilisée.
	if ms.Sys > ms.HeapInuse {
		return ms.Sys - ms.HeapInuse
	}
	return 0
}
