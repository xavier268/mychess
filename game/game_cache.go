package game

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xavier268/mychess/position"
)

// SearchDirs is the ordered list of directories where NewGame looks for a cache file.
// The first directory that contains a valid file wins; among all candidates in all
// directories the file with the highest fill percentage is chosen.
var SearchDirs = []string{".", "./bin", ".."}

// binary format constants — must stay in sync with cache/cache.go
var cacheMagic = [8]byte{'M', 'Y', 'C', 'H', 'C', 'A', 'C', 'H'}

const cacheFormatVersion = uint32(1)

// findBestCache scans dirs for cache files matching the current ZSize and returns
// the path and fill percentage of the one with the highest fill, or ("", 0, false).
func findBestCache(dirs []string) (path string, fillPct int, found bool) {
	prefix := fmt.Sprintf("mychess_cache_%dM_", ZSize/1_000_000)
	bestPct := -1

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".bin") {
				continue
			}
			pctStr := name[len(prefix) : len(name)-len(".bin")]
			pct, err := strconv.Atoi(pctStr)
			if err != nil || pct <= 0 || pct > 99 {
				continue
			}
			if pct > bestPct {
				bestPct = pct
				path = filepath.Join(dir, name)
			}
		}
	}

	if bestPct < 0 {
		return "", 0, false
	}
	return path, bestPct, true
}

// tryLoadCache attempts to load the best available cache from dirs.
// Returns the ZobristTable, ZMap, file path, and true on success;
// or zero values and false if no valid cache is found.
func tryLoadCache(dirs []string) (position.ZobristTable, *ZMap, string, bool) {
	path, _, found := findBestCache(dirs)
	if !found {
		return position.ZobristTable{}, nil, "", false
	}

	f, err := os.Open(path)
	if err != nil {
		return position.ZobristTable{}, nil, "", false
	}
	defer f.Close()

	br := bufio.NewReaderSize(f, 1<<16)

	var m [8]byte
	if _, err := io.ReadFull(br, m[:]); err != nil || m != cacheMagic {
		return position.ZobristTable{}, nil, "", false
	}

	var version uint32
	if err := binary.Read(br, binary.LittleEndian, &version); err != nil || version != cacheFormatVersion {
		return position.ZobristTable{}, nil, "", false
	}

	var zsize uint64
	if err := binary.Read(br, binary.LittleEndian, &zsize); err != nil || zsize != uint64(ZSize) {
		return position.ZobristTable{}, nil, "", false
	}

	var fp uint32 // fill percent stored in header (informational)
	if err := binary.Read(br, binary.LittleEndian, &fp); err != nil {
		return position.ZobristTable{}, nil, "", false
	}

	var zt position.ZobristTable
	if err := zt.Decode(br); err != nil {
		return position.ZobristTable{}, nil, "", false
	}

	z := NewZMap()
	if err := z.Decode(br); err != nil {
		return position.ZobristTable{}, nil, "", false
	}

	return zt, z, path, true
}
