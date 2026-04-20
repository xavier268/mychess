package game

import (
	"fmt"
	"mychess/position"
)

// DEBUG : reduced size for debugging
// const ZSize = 1_000 // mémoire volontairement contrainte pour tester la saturation de la table

type ZMap struct {
	// data storage
	data [ZSize]ZEntry
	// stats when trying to set or get data
	missedSet, hitSet int
	missedGet, hitGet int
	cellCount         int
}

func NewZMap() *ZMap {
	return &ZMap{}
}

// Reset hit/miss, not count.
func (z *ZMap) ResetStats() {
	z.missedSet, z.hitSet = 0, 0
	z.missedGet, z.hitGet = 0, 0
}

// Display stats about hit/miss ratio
func (z *ZMap) Stats() string {
	return fmt.Sprintf("Cells %dk/%dk(%2.1f%%)\nGet h:%dk(%2.1f%%) m:%dk(%2.1f%%) \nSet h:%dk(%2.1f%%) m:%dk(%2.1f%%)",
		z.cellCount/1_000, ZSize/1_000, 100.0*float64(z.cellCount)/ZSize,
		z.hitGet/1_000, 100.0*float64(z.hitGet)/float64(z.hitGet+z.missedGet), z.missedGet/1_000, 100.0*float64(z.missedGet)/float64(z.hitGet+z.missedGet),
		z.hitSet/1_000, 100.0*float64(z.hitSet)/float64(z.hitSet+z.missedSet), z.missedSet/1_000, 100.0*float64(z.missedSet)/float64(z.hitSet+z.missedSet))
}

type ZEntry struct {
	// Best move found until now (or null move)
	Best position.Move
	// Score (upper, lower, exact)
	Score position.Score
	// Confirm H to detect collisions (not 100% perfect, but risk of remaining collision is low)
	ConfirmH uint16
	// Depth of analysis at this stage
	Depth uint16
	// When was this entry last updated ?
	Age uint16
	// Score type : UPPER, LOWER, EXACT
	ScoreType ScoreType
}

type ScoreType uint8

const (
	FREE ScoreType = iota // mark an empty, available ZEntry cell
	UPPER
	LOWER
	EXACT
)

// Used to detect collisions (hash having the same rest modulo ZSize)
// We only use 16bits of the quotient, that should be enough to remove most collisions for large ZSize.
func ConfirmH(hash uint64) uint16 { return uint16(hash / ZSize) }

// Test if cell is empty and available for storage
func (ze ZEntry) IsEmpty() bool { return ze.ScoreType == FREE }

// Get a ZEntry cell, if it is available
// If not found, found = false.
func (z *ZMap) Get(h uint64) (ze ZEntry, found bool) {
	e := z.data[h%ZSize]
	if e.IsEmpty() || (e.ConfirmH != ConfirmH(h)) { // not found - either empty or contains another position
		z.missedGet++
		return ZEntry{}, false
	}
	// found
	z.hitGet++
	return e, true
}

// Try to Set a ZEntry.
// Updated returns true if the value was updated, false if no change was made.
func (z *ZMap) Set(h uint64, ze ZEntry) (updated bool) {
	// Compute index in the array
	i := h % ZSize
	ze.ConfirmH = ConfirmH(h)
	e := z.data[i]

	// If cell is free, just set it,
	// and update stats and cell count
	if e.IsEmpty() {
		z.hitSet++
		z.cellCount++
		z.data[i] = ze
		return true
	}

	// Now, cell is NOT free...
	// If cell is occupied with another value - hash collision detected
	if e.ConfirmH != ConfirmH(h) {
		// if same age or mor recent, ignore and do nothing
		if e.Age >= ze.Age {
			z.missedSet++
			return false
		} else {
			// if older, overwrite
			z.hitSet++
			z.data[i] = ze
			return true
		}
	}

	// Now, we are certain cell is occupied, with the same hash, for the same position.
	// Age does not matter anymore here
	// Compare depth : if new value has greater depth, overwrite
	if ze.Depth > e.Depth {
		z.hitSet++
		z.data[i] = ze
		return true
	} else { // old value has greater depth, only update age, ignore the rest.
		if e.Age < ze.Age { // stored value was older
			e.Age = ze.Age
			z.data[i] = e
			z.hitSet++
			return true
		} else { // stored value was younger, with better depth - do nothing
			z.missedSet++
			return false
		}
	}

}
