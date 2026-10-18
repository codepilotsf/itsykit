package sitefiles

import (
	"encoding/binary"
	"fmt"
)

// Native executable header validation. A declared package executable must be a
// native binary for the platform it is declared under, so a mislabeled shell
// script, an archive or a binary built for another architecture is refused
// before it is launched or shipped. Only standard-library parsing is used and
// callers pass a bounded header prefix, never the whole file.
//
// ELF is checked for the Unix systems Go targets; Mach-O (thin and fat) for
// darwin. Other platforms have no format constraint here and return nil.

var elfOS = map[string]bool{
	"linux": true, "android": true, "freebsd": true, "netbsd": true,
	"openbsd": true, "dragonfly": true, "solaris": true, "illumos": true,
}

// CheckNativeExecutable reports whether header begins a native executable for
// goos/goarch. header is the first bytes read from the file, at most 4 KiB.
func CheckNativeExecutable(header []byte, goos, goarch string) error {
	switch {
	case elfOS[goos]:
		return checkELF(header, goarch)
	case goos == "darwin":
		return checkMachO(header, goarch)
	default:
		return nil
	}
}

type elfTarget struct {
	class   byte
	machine uint16
}

var elfTargets = map[string]elfTarget{
	"386":      {1, 3},   // EM_386
	"amd64":    {2, 62},  // EM_X86_64
	"arm":      {1, 40},  // EM_ARM
	"arm64":    {2, 183}, // EM_AARCH64
	"mips":     {1, 8},   // EM_MIPS
	"mipsle":   {1, 8},
	"mips64":   {2, 8},
	"mips64le": {2, 8},
	"ppc64":    {2, 21}, // EM_PPC64
	"ppc64le":  {2, 21},
	"riscv64":  {2, 243}, // EM_RISCV
	"s390x":    {2, 22},  // EM_S390
	"loong64":  {2, 258}, // EM_LOONGARCH
}

func checkELF(header []byte, goarch string) error {
	if len(header) < 20 || header[0] != 0x7f || header[1] != 'E' || header[2] != 'L' || header[3] != 'F' {
		return fmt.Errorf("not an ELF executable")
	}
	want, ok := elfTargets[goarch]
	if !ok {
		return fmt.Errorf("unsupported architecture %s", goarch)
	}
	if header[4] != want.class {
		return fmt.Errorf("ELF class does not match %s", goarch)
	}
	var order binary.ByteOrder = binary.LittleEndian
	switch header[5] {
	case 1:
		order = binary.LittleEndian
	case 2:
		order = binary.BigEndian
	default:
		return fmt.Errorf("invalid ELF byte order")
	}
	if machine := order.Uint16(header[18:20]); machine != want.machine {
		return fmt.Errorf("ELF machine does not match %s", goarch)
	}
	return nil
}

// Mach-O CPU types from <mach/machine.h>.
const (
	cpuArchX86    = 7
	cpuArchX86_64 = 0x01000007
	cpuArchArm    = 12
	cpuArchArm64  = 0x0100000c
)

var machoTargets = map[string]uint32{
	"386":   cpuArchX86,
	"amd64": cpuArchX86_64,
	"arm":   cpuArchArm,
	"arm64": cpuArchArm64,
}

const (
	machMagic   = 0xfeedface
	machMagic64 = 0xfeedfacf
	machCigam   = 0xcefaedfe
	machCigam64 = 0xcffaedfe
	fatMagic    = 0xcafebabe
	fatCigam    = 0xbebafeca
)

func checkMachO(header []byte, goarch string) error {
	want, ok := machoTargets[goarch]
	if !ok {
		return fmt.Errorf("unsupported architecture %s", goarch)
	}
	if len(header) < 8 {
		return fmt.Errorf("not a Mach-O executable")
	}
	// The magic is always stored so that it reads on a little-endian host as one
	// of these values; thin CPU fields use the file's own byte order.
	magic := binary.LittleEndian.Uint32(header[:4])
	switch magic {
	case machMagic, machMagic64:
		if cpu := binary.LittleEndian.Uint32(header[4:8]); cpu != want {
			return fmt.Errorf("Mach-O CPU does not match %s", goarch)
		}
		return nil
	case machCigam, machCigam64:
		if cpu := binary.BigEndian.Uint32(header[4:8]); cpu != want {
			return fmt.Errorf("Mach-O CPU does not match %s", goarch)
		}
		return nil
	case fatMagic, fatCigam:
		return checkFatMachO(header, want)
	default:
		return fmt.Errorf("not a Mach-O executable")
	}
}

// checkFatMachO scans a universal binary's architecture table. The header stores
// its counts big-endian and each architecture entry is 20 bytes (fat) or 32
// bytes (fat64); only the leading cputype of each entry is needed.
func checkFatMachO(header []byte, want uint32) error {
	if len(header) < 8 {
		return fmt.Errorf("not a Mach-O executable")
	}
	magic := binary.BigEndian.Uint32(header[:4])
	entrySize := 20
	if magic == 0xcafebabf || magic == 0xbfbafeca {
		entrySize = 32
	}
	count := binary.BigEndian.Uint32(header[4:8])
	if count == 0 || count > 64 {
		return fmt.Errorf("invalid Mach-O architecture count")
	}
	for i := uint32(0); i < count; i++ {
		start := 8 + int(i)*entrySize
		if start+4 > len(header) {
			return fmt.Errorf("truncated Mach-O architecture table")
		}
		if binary.BigEndian.Uint32(header[start:start+4]) == want {
			return nil
		}
	}
	return fmt.Errorf("Mach-O architecture does not match")
}
