// Command gencache pre-computes a ZMap from the initial chess position and saves
// it as a binary cache file.  The cache can then be loaded at engine startup to
// make the transposition table immediately useful from the first move.
//
// Usage:
//
//	gencache [-dir <path>] [-fill <percent>]
//
// Flags:
//
//	-dir   Directory where the cache file is written (default: current directory).
//	-fill  Target fill percentage for the ZMap, 1–99 (default: 75).
//
// The output filename encodes both ZSize and the fill target so that different
// build configurations and fill levels coexist in the same directory.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xavier268/mychess/cache"
	"github.com/xavier268/mychess/game"
)

func main() {
	dir := flag.String("dir", ".", "directory to write the cache file")
	fillPct := flag.Int("fill", 75, "target fill percentage (1–99)")
	flag.Parse()

	if *fillPct < 1 || *fillPct > 99 {
		fmt.Fprintln(os.Stderr, "error: fill must be between 1 and 99")
		os.Exit(1)
	}

	path := cache.FileName(*dir, *fillPct)
	fmt.Printf("ZSize   : %d entries\n", game.ZSize)
	fmt.Printf("Target  : %d%% fill\n", *fillPct)
	fmt.Printf("Output  : %s\n\n", path)

	start := time.Now()
	err := cache.Generate(*dir, *fillPct, func(fill float64) {
		fmt.Printf("\r  fill: %5.1f%%   elapsed: %v   ", fill, time.Since(start).Round(time.Second))
	})
	fmt.Println() // newline after progress line
	if err != nil {
		log.Fatalf("generation failed: %v", err)
	}
	fmt.Printf("Done in %v\n", time.Since(start).Round(time.Second))
}
