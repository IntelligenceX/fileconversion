package ole2

import (
	"io"
	"log"
)

// DEBUG enables debug
var DEBUG = false

// StreamReader is an OLE stream reader
type StreamReader struct {
	sat            []int32
	start          int32
	reader         io.ReadSeeker
	offsetOfSector int32
	offsetInSector uint32
	sizeSector     uint32
	size           int64
	offset         int64
	sectorPos      func(int32, uint32) int32
}

// Read reads data from the stream into p
func (r *StreamReader) Read(p []byte) (n int, err error) {
	if r.offsetOfSector == EOFSecID {
		return 0, io.EOF
	}
	pos := r.sectorPos(r.offsetOfSector, r.sizeSector) + int32(r.offsetInSector)
	r.reader.Seek(int64(pos), 0)
	readed := uint32(0)
	for remainLen := uint32(len(p)) - readed; remainLen > r.sizeSector-r.offsetInSector; remainLen = uint32(len(p)) - readed {
		if n, err := r.reader.Read(p[readed : readed+r.sizeSector-r.offsetInSector]); err != nil {
			return int(readed) + n, err
		} else {
			readed += uint32(n)
			r.offsetInSector = 0
			if r.offsetOfSector >= int32(len(r.sat)) {
				//log.Fatal(`
				//THIS SHOULD NOT HAPPEN, IF YOUR PROGRAM BREAK,
				//COMMENT THIS LINE TO CONTINUE AND MAIL ME XLS FILE
				//TO TEST, THANKS`)
				return int(readed), io.EOF
			} else {
				r.offsetOfSector = r.sat[r.offsetOfSector]
			}
			if r.offsetOfSector == EOFSecID {
				return int(readed), io.EOF
			}
			pos := r.sectorPos(r.offsetOfSector, r.sizeSector) + int32(r.offsetInSector)
			r.reader.Seek(int64(pos), 0)
		}
	}
	if n, err := r.reader.Read(p[readed:len(p)]); err == nil {
		r.offsetInSector += uint32(n)
		if DEBUG {
			log.Printf("pos:%x,bit:% X", r.offsetOfSector, p)
		}
		return len(p), nil
	} else {
		return int(readed) + n, err
	}

}

// Seek seeks the stream to the given offset
func (r *StreamReader) Seek(offset int64, whence int) (offsetResult int64, err error) {

	if whence == 0 {
		r.offsetOfSector = r.start
		r.offsetInSector = 0
		r.offset = offset
	} else {
		r.offset += offset
	}

	if r.offsetOfSector == EOFSecID {
		return r.offset, io.EOF
	}

	for offset >= int64(r.sizeSector-r.offsetInSector) {
		r.offsetOfSector = r.sat[r.offsetOfSector]
		offset -= int64(r.sizeSector - r.offsetInSector)
		r.offsetInSector = 0
		if r.offsetOfSector == EOFSecID {
			err = io.EOF
			goto return_res
		}
	}

	if r.size <= r.offset {
		err = io.EOF
		r.offset = r.size
	} else {
		r.offsetInSector += uint32(offset)
	}
return_res:
	offsetResult = r.offset
	return
}
