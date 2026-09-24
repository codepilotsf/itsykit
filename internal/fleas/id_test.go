package fleas

import (
	"bytes"
	"strings"
	"testing"
)

func TestReserveID(t *testing.T) {
	used := map[string]struct{}{}
	for i := 0; i < 1000; i++ {
		id, err := ReserveID(used)
		if err != nil || len(id) != IDLength || len(used) != i+1 {
			t.Fatal(id, err, len(used))
		}
		for _, ch := range id {
			if !strings.ContainsRune(alphabet, ch) {
				t.Fatalf("invalid ID %q", id)
			}
		}
	}
}

func TestCollisionRetriesWithoutReplacingExistingID(t *testing.T) {
	used := map[string]struct{}{"aaaaaaaaaa": {}}
	entropy := append(make([]byte, 10), bytes.Repeat([]byte{1}, 10)...)
	id, err := reserveID(bytes.NewReader(entropy), used)
	if err != nil || id != "bbbbbbbbbb" || len(used) != 2 {
		t.Fatal(id, err, used)
	}
}

func TestRejectBiasedBytes(t *testing.T) {
	entropy := append([]byte{248, 249, 255}, []byte{0, 25, 26, 51, 52, 61, 62, 123, 186, 247}...)
	id, err := reserveID(bytes.NewReader(entropy), map[string]struct{}{})
	if err != nil || id != "azAZ09a9a9" {
		t.Fatal(id, err)
	}
}

func TestAllocationFailureDoesNotReserve(t *testing.T) {
	for _, entropy := range [][]byte{nil, {1, 2}, bytes.Repeat([]byte{255}, 1024), make([]byte, 1280)} {
		used := map[string]struct{}{"aaaaaaaaaa": {}}
		if _, err := reserveID(bytes.NewReader(entropy), used); err == nil || len(used) != 1 {
			t.Fatal(err, used)
		}
	}
	if _, err := reserveID(strings.NewReader(""), nil); err == nil {
		t.Fatal("accepted absent collision index")
	}
}
