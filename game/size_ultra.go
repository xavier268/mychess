//go:build ultra

package game

// Size of ZEntry storage (nb of entries)
// ultra	500M d'entrées		25Go
// high		50M d'entrées		4Go
// default	5M d'entrées		400Mo
// low		1M d'entrées		90Mo
const ZSize = 500_000_000
