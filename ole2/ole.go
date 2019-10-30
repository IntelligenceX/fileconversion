package ole2

import (
	"encoding/binary"
	"io"
)

// SecIDType
const (
	MSATSecID = iota - 4
	SATSecID
	EOFSecID
	FreeSecID
)

// Ole represents an OLE file
type Ole struct {
	header   *Header
	Lsector  uint32
	Lssector uint32
	SecID    []int32
	SSecID   []int32
	Files    []File
	reader   io.ReadSeeker
}

func Open(reader io.ReadSeeker, charset string) (ole *Ole, err error) {
	var header *Header
	var hbts = make([]byte, 512)
	reader.Read(hbts)
	if header, err = parseHeader(hbts); err == nil {
		ole = new(Ole)
		ole.reader = reader
		ole.header = header
		ole.Lsector = 512 //TODO
		ole.Lssector = 64 //TODO

		if err = ole.readMSAT(); err != nil {
			return ole, err
		}

		if err = ole.readSSAT(); err != nil {
			return ole, err
		}

		return ole, nil
	}

	return nil, err
}

func (o *Ole) ListDir() (dir []*File, err error) {
	sector := o.stream_read(o.header.FirstSecIDDirectory, 0)
	dir = make([]*File, 0)
	for {
		d := new(File)
		err = binary.Read(sector, binary.LittleEndian, d)
		if err == nil && d.Type != EMPTY {
			dir = append(dir, d)
		} else {
			break
		}
	}
	if err == io.EOF && dir != nil {
		return dir, nil
	}

	return
}

func (o *Ole) OpenFile(file *File, root *File) io.ReadSeeker {
	if file.Size < o.header.MinSizeOfStandardStream {
		return o.short_stream_read(file.Sstart, file.Size, root.Sstart)
	} else {
		return o.stream_read(file.Sstart, file.Size)
	}
}

// Read MSAT
func (o *Ole) readMSAT() error {
	for i := uint32(0); i < 109 && i < o.header.NumberOfSectorsSAT; i++ {
		if sector, err := o.sector_read(o.header.FirstPartOfMSAT[i]); err == nil {
			sids := sector.AllValues(o.Lsector)
			o.SecID = append(o.SecID, sids...)
		} else {
			return err
		}
	}

	if o.header.NumberOfSectorsSAT > 109 && o.header.NumberOfSectorsMSAT != 0 && o.header.FirstSecIDMSAT >= 0 {
		sid := o.header.FirstSecIDMSAT

		for j := uint32(109); sid != EOFSecID && j < o.header.NumberOfSectorsSAT; {
			if sector, err := o.sector_read(sid); err == nil {
				sids := sector.MsatValues(o.Lsector)

				for _, sid := range sids {
					j++
					if sector, err := o.sector_read(int32(sid)); err == nil {
						sids := sector.AllValues(o.Lsector)

						o.SecID = append(o.SecID, sids...)
					} else {
						return err
					}
				}

				sid = sector.NextSid(o.Lsector)
			} else {
				return err
			}
		}
	}

	return nil
}

func (o *Ole) readSSAT() error {
	sid := o.header.FirstSecIDSSAT

	for i := uint32(0); i < o.header.NumberOfSectorsSSAT; i++ {
		if sid != EOFSecID {
			if sector, err := o.sector_read(sid); err == nil {
				sids := sector.MsatValues(o.Lsector)

				o.SSecID = append(o.SSecID, sids...)

				sid = sector.NextSid(o.Lsector)
			} else {
				return err
			}
		}
	}

	return nil
}

func (o *Ole) stream_read(sid int32, size uint32) *StreamReader {
	return &StreamReader{o.SecID, sid, o.reader, sid, 0, o.Lsector, int64(size), 0, sector_pos}
}

func (o *Ole) short_stream_read(sid int32, size uint32, startSecId int32) *StreamReader {
	ssatReader := &StreamReader{o.SecID, startSecId, o.reader, sid, 0, o.Lsector, int64(uint32(len(o.SSecID)) * o.Lssector), 0, sector_pos}
	return &StreamReader{o.SSecID, sid, ssatReader, sid, 0, o.Lssector, int64(size), 0, short_sector_pos}
}

func (o *Ole) sector_read(sid int32) (Sector, error) {
	return o.sector_read_internal(sid, o.Lsector)
}

func (o *Ole) short_sector_read(sid int32) (Sector, error) {
	return o.sector_read_internal(sid, o.Lssector)
}

func (o *Ole) sector_read_internal(sid int32, size uint32) (Sector, error) {
	pos := sector_pos(sid, size)
	if _, err := o.reader.Seek(int64(pos), 0); err == nil {
		var bts = make([]byte, size)
		o.reader.Read(bts)
		return Sector(bts), nil
	} else {
		return nil, err
	}
}

func sector_pos(sid int32, size uint32) int32 {
	return 512 + sid*int32(size)
}

func short_sector_pos(sid int32, size uint32) int32 {
	return sid * int32(size)
}
