# fileconversion

This is a Go library to convert various file formats into plaintext and provide related useful functions.

This library is used for https://intelx.io and was successfully tested over 184 million individual files. It is partly written from scratch, partly forked from open source and partly a rewrite of existing code. Many existing libraries lack stability and functionality and this libraries solves that. 

We welcome any contributions - please open issues for any feature requests, bugs, and other related issues.

It supports following file formats for plaintext conversion:

* Word: DOC, DOCX, RTF, ODT
* Excel: XLS, XLSX, ODS
* PowerPoint: PPTX
* PDF
* Ebook: EPUB, MOBI
* Website: HTML

Functions for compressed and container files:

* Decompress files: GZ, BZ, BZ2, XZ
* Extract files from containers: ZIP, RAR, 7Z, TAR

Picture related functions:

* Check if pictures are excessively large
* Compress (and convert) pictures to JPEG: GIF, JPEG, PNG, BMP, TIFF
* Resize and compress pictures
* Extract pictures from PDF files

To download this library:

```
go get -u github.com/IntelligenceX/fileconversion
```

And then use it like:

```go
package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/IntelligenceX/fileconversion"
)

const sizeLimit = 2 * 1024 * 1024 // 2 MB

func main() {
	// extract text from an XLSX file
	file, err := os.Open("Test.xlsx")
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return
	}

	defer file.Close()
	stat, _ := file.Stat()

	buffer := bytes.NewBuffer(make([]byte, 0, sizeLimit))

	fileconversion.XLSX2Text(file, stat.Size(), buffer, sizeLimit, -1)

	fmt.Println(buffer.String())
}
```


## Functions

The package exports the following functions:

```go
XLSX2Text(file io.ReaderAt, size int64, writer io.Writer, limit int64, rowLimit int) (written int64, err error)
DOCX2Text(file io.ReaderAt, size int64) (string, error)
EPUB2Text(file io.ReaderAt, size int64, limit int64) (string, error)
HTML2Text(reader io.Reader) (pageText string, err error)
HTML2TextAndLinks(reader io.Reader, baseURL string) (pageText string, links []string, err error)
Mobi2Text(file io.ReadSeeker) (string, error)
ODS2Text(file io.ReaderAt, size int64, writer io.Writer, limit int64) (written int64, err error)
ODT2Text(file io.ReaderAt, size int64, writer io.Writer, limit int64) (written int64, err error)
PDFListContentStreams(f io.ReadSeeker, w io.Writer, size int64) (written int64, err error)
PPTX2Text(file io.ReaderAt, size int64) (string, error)
RTF2Text(inputRtf string) string
XLS2Text(reader io.ReadSeeker, writer io.Writer, size int64) (written int64, err error)
XLSX2Text(file io.ReaderAt, size int64, writer io.Writer, limit int64, rowLimit int) (written int64, err error)
```

Picture functions:

```go
IsExcessiveLargePicture(Picture []byte) (excessive bool, err error)
CompressJPEG(Picture []byte, quality int) (compressed []byte)
ResizeCompressPicture(Picture []byte, Quality int, MaxWidth, MaxHeight uint) 
PDFExtractImages(input io.ReadSeeker) (images []ImageResult, err error)
```

Compression and container file functions:

```go
DecompressFile(data []byte) (decompressed []byte, valid bool)
ContainerExtractFiles(data []byte, callback func(name string, size int64, date time.Time, data []byte))
```

## Dependencies

This library uses other go packages. Run the following command to download them:

```
go get -u github.com/nwaples/rardecode
go get -u github.com/saracen/go7z
go get -u github.com/ulikunitz/xz
go get -u github.com/mattetti/filebuffer
go get -u github.com/richardlehane/mscfb
go get -u github.com/taylorskalyo/goreader/epub
go get -u github.com/PuerkitoBio/goquery
go get -u github.com/ssor/bom
go get -u github.com/levigross/exp-html
go get -u github.com/neofight/mobi/convert
go get -u github.com/neofight/mobi/headers
go get -u github.com/unidoc/unipdf
go get -u github.com/nfnt/resize
go get -u github.com/tealeg/xlsx
go get -u gopkg.in/xmlpath.v2
```

## Tests

There are no functional tests. The only test functions are used manually for debugging.

## Forks

Other packages were tested and either found insufficient, or unstable. Many of the below listed packages were found to be unstable, cause crashes, as well as exhaust memory due to bad programming, bad input sanitizing and bad memory management.

* `html2text` is forked from https://github.com/jaytaylor/html2text
* `odf` is forked from https://github.com/knieriem/odf
* `ole2` is forked and partly rewritten from https://github.com/extrame/ole2
* `xls` is forked from https://github.com/sergeilem/xls which is a fork from https://github.com/extrame/xls
* `doc` is forked from https://github.com/EndFirstCorp/doc2txt
* `docx` is forked from https://github.com/guylaor/goword
* `mobi` is forked from https://github.com/neofight/mobi
* `odt` is forked from https://github.com/lu4p/cat
* `pptx` is forked from https://github.com/mr-tim/rol-o-decks
* `rtf` is forked from https://github.com/J45k4/rtf-go

## License

This is free and unencumbered software released into the public domain.

Note that this package includes, or consists partly of forks or rewrite of existing open source code. Use at your own risk. Intelligence X does not provide any warranty for this library or any parts of it.
