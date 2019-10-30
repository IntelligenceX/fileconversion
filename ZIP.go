/*
File Name:  ZIP.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner
*/

package fileconversion

import "bytes"

// IsFileZIP checks if the data indicates a ZIP file.
// Many file formats like DOCX, XLSX, PPTX and APK are actual ZIP files.
// Signature 50 4B 03 04
func IsFileZIP(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04})
}
