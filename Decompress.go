/*
File Name:  Decompress.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner
*/

package fileconversion

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"io/ioutil"
	"time"

	"github.com/nwaples/rardecode"
	"github.com/saracen/go7z"
	"github.com/ulikunitz/xz"
)

// DecompressFile decompresses data. It supports: GZ, BZ, BZ2, XZ
func DecompressFile(data []byte) (decompressed []byte, valid bool) {
	// Try GZ
	if gr, err := gzip.NewReader(bytes.NewBuffer(data)); err == nil {
		defer gr.Close()
		decompressed, err = ioutil.ReadAll(gr)
		if err == nil {
			return decompressed, true
		}
	}

	// BZ, BZ2
	br := bzip2.NewReader(bytes.NewBuffer(data))
	decompressed, err := ioutil.ReadAll(br)
	if err == nil {
		return decompressed, true
	}

	// XZ
	if xr, err := xz.NewReader(bytes.NewBuffer(data)); err == nil {
		decompressed, err = ioutil.ReadAll(xr)
		if err == nil {
			return decompressed, true
		}
	}

	return nil, false
}

// ContainerExtractFiles extracts files from supported containers: ZIP, RAR, 7Z, TAR
func ContainerExtractFiles(data []byte, callback func(name string, size int64, date time.Time, data []byte)) {

	// ZIP
	if r, err := zip.NewReader(bytes.NewReader(data), int64(len(data))); err == nil {
		for _, f := range r.File {
			fileReader, err := f.Open()
			if err != nil {
				continue
			}

			data2, err := ioutil.ReadAll(fileReader)
			fileReader.Close()
			if err != nil {
				// If the file is encrypted with a password, this fails with error "4" here.
				continue
			}

			callback(f.Name, int64(f.UncompressedSize64), f.Modified, data2)
		}

		return
	}

	// RAR
	if rc, err := rardecode.NewReader(bytes.NewReader(data), ""); err == nil {
		for {
			hdr, err := rc.Next()
			if err == io.EOF || err != nil { // break if end of archive or other error returned
				break
			} else if err == nil && !hdr.IsDir {
				if data2, err := ioutil.ReadAll(rc); err == nil {
					callback(hdr.Name, hdr.UnPackedSize, hdr.CreationTime, data2)
				}
			}
		}
	}

	// 7Z
	if sz, err := go7z.NewReader(bytes.NewReader(data), int64(len(data))); err == nil {
		for {
			hdr, err := sz.Next()
			if err == io.EOF || err != nil { // break if end of archive or other error returned
				break // End of archive
			} else if err == nil && !hdr.IsEmptyFile {
				if data2, err := ioutil.ReadAll(sz); err == nil {
					callback(hdr.Name, int64(len(data2)), hdr.CreatedAt, data2)
				}
			}
		}
	} else if err == go7z.ErrDecompressorNotFound {
		// May happen if it's 7Z, but decompressor not available (like 7zAES).
		return
	}

	// TAR
	tr := tar.NewReader(bytes.NewReader(data))
	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			// other error
			break
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			// directories are ignored
		case tar.TypeReg, tar.TypeRegA:
			// file
			data2, err := ioutil.ReadAll(tr)
			if err != nil {
				continue
			}

			callback(hdr.Name, hdr.Size, hdr.ModTime, data2)
		}
	}

}
