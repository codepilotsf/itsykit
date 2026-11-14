package site

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Opt-in comparison probe: run identically against the baseline and new loader.
// Uses a sparse 16 MiB asset; measures live Go heap, not OS cache or total RSS.
func TestAssetMemoryProbe(t *testing.T) {
	if os.Getenv("FLEA_ASSET_MEMORY_PROBE") != "1" {
		t.Skip("set FLEA_ASSET_MEMORY_PROBE=1 to measure retained heap")
	}
	root := fixture(t)
	f, err := os.Create(filepath.Join(root, "site/large.bin"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Truncate(16 << 20)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	p := mustLoad(t, root)
	runtime.GC()
	runtime.ReadMemStats(&after)
	var retained int
	for _, data := range p.AssetData {
		retained += len(data)
	}
	t.Logf("asset bytes retained=%d; live heap delta=%d; allocated during load=%d", retained, int64(after.HeapAlloc)-int64(before.HeapAlloc), after.TotalAlloc-before.TotalAlloc)
	runtime.KeepAlive(p)
}
