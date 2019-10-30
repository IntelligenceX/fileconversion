package ole2

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Header represents the OLE header
type Header struct {
	Signature               [2]uint32
	Clid                    [4]uint32
	RevisionNumber          uint16
	VersionNumber           uint16
	ByteOrder               uint16
	SizeOfSector            uint16
	SizeOfShortSector       uint16
	_                       uint16
	_                       uint64
	NumberOfSectorsSAT      uint32 //Total number of sectors used for the sector allocation table
	FirstSecIDDirectory     int32  //SecID of first sector of the directory stream
	_                       uint32
	MinSizeOfStandardStream uint32 //Minimum size of a standard stream
	FirstSecIDSSAT          int32  //SecID of first sector of the short-sector allocation table
	NumberOfSectorsSSAT     uint32 //Total number of sectors used for the short-sector allocation table
	FirstSecIDMSAT          int32  //SecID of first sector of the master sector allocation table
	NumberOfSectorsMSAT     uint32 //Total number of sectors used for the master sector allocation table
	FirstPartOfMSAT         [109]int32
}

func parseHeader(bts []byte) (*Header, error) {
	buf := bytes.NewBuffer(bts)
	header := new(Header)
	binary.Read(buf, binary.LittleEndian, header)
	if header.Signature[0] != 0xE011CFD0 || header.Signature[1] != 0xE11AB1A1 || header.ByteOrder != 0xFFFE {
		return nil, fmt.Errorf("not an excel file")
	}

	return header, nil
}
