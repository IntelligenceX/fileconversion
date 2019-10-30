/*
File Name:  PPT 2 Text.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner

Placeholder file until PPT conversion code 2 text is available.
*/

package fileconversion

import "bytes"

// IsFilePPT checks if the data indicates a PPT file
// PPT has multiple signature according to https://www.filesignatures.net/index.php?page=search&search=PPT&mode=EXT, D0 CF 11 E0 A1 B1 1A E1. This overlaps with others (including DOC ans XLS).
func IsFilePPT(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
}
