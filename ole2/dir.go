package ole2

import (
	"unicode/utf16"
)

// constants for directory
const (
	EMPTY       = iota
	USERSTORAGE = iota
	USERSTREAM  = iota
	LOCKBYTES   = iota
	PROPERTY    = iota
	ROOT        = iota
)

// File is an OLE file
type File struct {
	NameBts   [32]uint16
	Bsize     uint16
	Type      byte
	Flag      byte
	Left      uint32
	Right     uint32
	Child     uint32
	GUID      [8]uint16
	Userflags uint32
	Time      [2]uint64
	Sstart    int32
	Size      uint32
	Proptype  uint32
}

// Name returns the file name
func (d *File) Name() string {
	if int(d.Bsize)/2-1 > len(d.NameBts) || int(d.Bsize)/2-1 < 0 {
		return ""
	}

	runes := utf16.Decode(d.NameBts[:d.Bsize/2-1])
	return string(runes)
}
