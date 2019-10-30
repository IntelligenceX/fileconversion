/*
File Name:  PDF 2 Text.go
Copyright:  2018 Kleissner Investments s.r.o.
Author:     Peter Kleissner

This code uses the commercial library UniDoc https://unidoc.io/ to extract text from PDFs.
*/

package fileconversion

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/unidoc/unipdf/core"
	"github.com/unidoc/unipdf/extractor"
	pdf "github.com/unidoc/unipdf/model"

	"github.com/unidoc/unipdf/common/license"
)

// InitPDFLicense initializes the PDF license
func InitPDFLicense(key, name string) {
	// load the unidoc license (v3)
	license.SetLicenseKey(key, name)
}

// PDFListContentStreams writes all text streams in a PDF to the writer
// It returns the number of characters attempted written (excluding "Page N" and new-lines) and an error, if any. It can be used to determine whether any text was extracted.
// The parameter size is the max amount of bytes (not characters) to write out.
func PDFListContentStreams(f io.ReadSeeker, w io.Writer, size int64) (written int64, err error) {

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return 0, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return 0, err
	}

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return 0, err
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return 0, err
	}

	for i := 0; i < numPages && size > 0; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return written, err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return written, err
		}

		txt, err := ex.ExtractText()
		if err != nil {
			return written, err
		}

		// use the extracted text
		txtNL := ""
		if written > 0 {
			txtNL += "\n\n"
		}

		textB := []byte(txtNL + "---- Page " + strconv.Itoa(pageNum) + " ----\n")

		// empty page? skip if so.
		txt = strings.TrimSpace(txt)
		if len(txt) == 0 {
			continue
		}

		textB = append(textB, []byte(txt)...)
		if int64(len(textB)) > size {
			textB = textB[:size]
		}

		if _, err = w.Write(textB); err != nil {
			return written, err
		}

		size -= int64(len(textB))
		written += int64(len(txt))
	}

	return written, nil
}

// PDFGetCreationDate tries to get the creation date
func PDFGetCreationDate(f io.ReadSeeker) (date time.Time, valid bool) {
	// Below code is forked from https://github.com/unidoc/unidoc-examples/blob/master/pdf/metadata/pdf_metadata_get_docinfo.go
	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return date, false
	}

	trailerDict, err := pdfReader.GetTrailer()
	if err != nil || trailerDict == nil {
		return date, false
	}

	var infoDict *core.PdfObjectDictionary

	infoObj := trailerDict.Get("Info")
	switch t := infoObj.(type) {
	case *core.PdfObjectReference:
		infoRef := t
		infoObj, err = pdfReader.GetIndirectObjectByNumber(int(infoRef.ObjectNumber))
		infoObj = core.TraceToDirectObject(infoObj)
		if err != nil {
			return date, false
		}
		infoDict, _ = infoObj.(*core.PdfObjectDictionary)
	case *core.PdfObjectDictionary:
		infoDict = t
	}

	if infoDict == nil {
		return date, false
	}

	if str, has := infoDict.Get("CreationDate").(*core.PdfObjectString); has {
		creationDateA := strings.TrimPrefix(str.String(), "D:")

		time1, err := time.Parse("20060102150405-07'00'", creationDateA)
		if err == nil {
			return time1.UTC(), true
		}
	}

	return date, false
}
