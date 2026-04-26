// Package cache persists a ZobristTable and a pre-computed ZMap to disk so that
// the engine can start with a warm transposition table from the very first move.
//
// File naming: mychess_cache_<ZSize>_<fillPct>.bin
// One file per memory configuration (ZSize varies with build tags low/high/ultra).
//
// File format (little-endian):
//
//	[8]byte  magic        "MYCHCACH"
//	uint32   version      format version (currently 1)
//	uint64   ZSize        entries in the table (build-tag dependent)
//	uint32   fillPct      target fill percentage used during generation
//	[N]byte  ZobristTable encoding/binary serialisation of the table
//	[M]byte  ZMap.data    raw unsafe dump of the [ZSize]ZEntry array
//	int64    cellCount    number of populated entries
package cache

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/xavier268/mychess"
	"github.com/xavier268/mychess/game"
	"github.com/xavier268/mychess/position"
)

const formatVersion = uint32(1)

// FileName returns the canonical cache file path for the given directory and
// fill percentage.
func FileName(dir string, fillPct int) string {
	return filepath.Join(dir, fmt.Sprintf("mychess_cache_%dM_%02d.bin", game.ZSize/1_000_000, fillPct))
}

// Generate runs alpha-beta search from the initial position until the ZMap
// reaches fillPct% capacity, then writes the cache file to dir.
// onProgress is called roughly every second with the current fill percentage;
// pass nil to disable.
func Generate(dir string, fillPct int, onProgress func(fill float64)) error {
	g := game.NewGame()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fill := g.Z.FillPercent()
				if onProgress != nil {
					onProgress(fill)
				}
				if fill >= float64(fillPct) {
					cancel()
					return
				}
			}
		}
	}()

	g.Analysis(ctx, 99)
	cancel() // stop the monitor goroutine (idempotent)

	return writeFile(FileName(dir, fillPct), fillPct, g.Z, &position.DefaultZT)
}

func writeFile(path string, fillPct int, z *game.ZMap, zt *position.ZobristTable) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 1<<20) // 1 MB write buffer

	if _, err := bw.Write(mychess.CacheMagic[:]); err != nil {
		return err
	}
	if err := binary.Write(bw, binary.LittleEndian, formatVersion); err != nil {
		return err
	}
	if err := binary.Write(bw, binary.LittleEndian, uint64(game.ZSize)); err != nil {
		return err
	}
	if err := binary.Write(bw, binary.LittleEndian, uint32(fillPct)); err != nil {
		return err
	}
	if err := zt.Encode(bw); err != nil {
		return err
	}
	if err := z.Encode(bw); err != nil {
		return err
	}
	return bw.Flush()
}

// Load reads the cache file for the given directory and fill percentage.
// Returns the stored ZobristTable and ZMap.
// Call position.RestoreDefaultZT(zt) before using the ZMap to ensure
// hash consistency.
func Load(dir string, fillPct int) (position.ZobristTable, *game.ZMap, error) {
	f, err := os.Open(FileName(dir, fillPct))
	if err != nil {
		return position.ZobristTable{}, nil, err
	}
	defer f.Close()

	br := bufio.NewReaderSize(f, 1<<16) // 64 KB read buffer

	// Validate magic
	var m [8]byte
	if _, err := io.ReadFull(br, m[:]); err != nil {
		return position.ZobristTable{}, nil, fmt.Errorf("reading magic: %w", err)
	}
	if m != mychess.CacheMagic {
		return position.ZobristTable{}, nil, fmt.Errorf("not a mychess cache file")
	}

	// Validate format version
	var version uint32
	if err := binary.Read(br, binary.LittleEndian, &version); err != nil {
		return position.ZobristTable{}, nil, fmt.Errorf("reading version: %w", err)
	}
	if version != formatVersion {
		return position.ZobristTable{}, nil, fmt.Errorf("unsupported cache version %d", version)
	}

	// Validate ZSize
	var zsize uint64
	if err := binary.Read(br, binary.LittleEndian, &zsize); err != nil {
		return position.ZobristTable{}, nil, fmt.Errorf("reading ZSize: %w", err)
	}
	if zsize != uint64(game.ZSize) {
		return position.ZobristTable{}, nil, fmt.Errorf("ZSize mismatch: file has %d, this build expects %d", zsize, game.ZSize)
	}

	// Read fill percent (informational, not validated)
	var fp uint32
	if err := binary.Read(br, binary.LittleEndian, &fp); err != nil {
		return position.ZobristTable{}, nil, fmt.Errorf("reading fillPct: %w", err)
	}

	// Read ZobristTable
	var zt position.ZobristTable
	if err := zt.Decode(br); err != nil {
		return position.ZobristTable{}, nil, fmt.Errorf("reading ZobristTable: %w", err)
	}

	// Read ZMap
	z := game.NewZMap()
	if err := z.Decode(br); err != nil {
		return position.ZobristTable{}, nil, fmt.Errorf("reading ZMap: %w", err)
	}

	return zt, z, nil
}
