package odf

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"strings"
)

const (
	MimeTypePfx = "application/vnd.oasis.opendocument."
)

type File struct {
	*zip.Reader
	cl       io.Closer
	MimeType string
}

// Open an OpenDocument file for reading, and check its MIME type.
// The returned *File provides -- via its Open method -- access to
// files embedded in the ODF, like content.xml.
func Open(odfName string) (*File, error) {
	z, err := zip.OpenReader(odfName)
	if err != nil {
		return nil, err
	}
	return newFile(&z.Reader, z)
}

// NewReader initializes a File struct with an already opened ODF
// file, and checks the file's MIME type. The returned *File provides
// access to files embedded in the ODF file, like content.xml.
func NewReader(r io.ReaderAt, size int64) (*File, error) {
	z, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	return newFile(z, nil)
}

func newFile(z *zip.Reader, closer io.Closer) (*File, error) {
	f := new(File)
	f.Reader = z
	mf, err := f.Open("mimetype")
	if err != nil {
		if closer != nil {
			closer.Close()
		}
		return nil, err
	}

	b, err := ioutil.ReadAll(mf)
	mf.Close()
	if err != nil {
		if closer != nil {
			closer.Close()
		}
		return nil, err
	}
	f.MimeType = string(b)
	f.cl = closer

	if !strings.HasPrefix(f.MimeType, MimeTypePfx) {
		return nil, errors.New("not an Open Document mime type")
	}
	return f, nil
}

func (f *File) Close() error {
	if f.cl == nil {
		return nil
	}
	return f.cl.Close()
}

func (f *File) Open(name string) (io.ReadCloser, error) {
	for _, zf := range f.File {
		if zf.Name == name {
			return zf.Open()
		}
	}
	return nil, errors.New("odf: open " + name + ": no such file")
}
