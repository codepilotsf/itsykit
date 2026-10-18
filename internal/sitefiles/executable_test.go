package sitefiles

import (
	"encoding/binary"
	"testing"
)

func elfHeader(class byte, machine uint16) []byte {
	h := make([]byte, 64)
	copy(h, "\x7fELF")
	h[4], h[5] = class, 1
	binary.LittleEndian.PutUint16(h[18:], machine)
	return h
}

func machoHeader(cpu uint32) []byte {
	h := make([]byte, 32)
	binary.LittleEndian.PutUint32(h[0:], 0xfeedfacf)
	binary.LittleEndian.PutUint32(h[4:], cpu)
	return h
}

func fatMacho(cpus ...uint32) []byte {
	const entry = 20
	h := make([]byte, 8+entry*len(cpus))
	binary.BigEndian.PutUint32(h[0:], 0xcafebabe)
	binary.BigEndian.PutUint32(h[4:], uint32(len(cpus)))
	for i, cpu := range cpus {
		binary.BigEndian.PutUint32(h[8+i*entry:], cpu)
	}
	return h
}

func TestCheckNativeExecutable(t *testing.T) {
	if err := CheckNativeExecutable(elfHeader(2, 62), "linux", "amd64"); err != nil {
		t.Fatalf("linux/amd64 ELF rejected: %v", err)
	}
	if err := CheckNativeExecutable(elfHeader(2, 183), "linux", "arm64"); err != nil {
		t.Fatalf("linux/arm64 ELF rejected: %v", err)
	}
	if err := CheckNativeExecutable(machoHeader(0x0100000c), "darwin", "arm64"); err != nil {
		t.Fatalf("darwin/arm64 Mach-O rejected: %v", err)
	}
	if err := CheckNativeExecutable(fatMacho(7, 0x0100000c), "darwin", "arm64"); err != nil {
		t.Fatalf("fat Mach-O rejected: %v", err)
	}
	// Platforms outside the native-format set have no constraint.
	if err := CheckNativeExecutable([]byte("MZ"), "windows", "amd64"); err != nil {
		t.Fatalf("unexpected constraint: %v", err)
	}

	for _, tc := range []struct {
		name   string
		header []byte
		goos   string
		goarch string
	}{
		{"script", []byte("#!/bin/sh\n"), "linux", "amd64"},
		{"wrong machine", elfHeader(2, 183), "linux", "amd64"},
		{"wrong class", elfHeader(1, 62), "linux", "amd64"},
		{"short", []byte{0x7f, 'E'}, "linux", "amd64"},
		{"macho wrong cpu", machoHeader(0x01000007), "darwin", "arm64"},
		{"macho not macho", elfHeader(2, 62), "darwin", "arm64"},
		{"fat missing arch", fatMacho(7), "darwin", "arm64"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckNativeExecutable(tc.header, tc.goos, tc.goarch); err == nil {
				t.Fatal("invalid native executable accepted")
			}
		})
	}
}
