/*
File Name:  XLSX 2 Text.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner

* https://github.com/tealeg/xlsx is used in production.
Some files used more than 1 GB of memory, even though the file itself is only 9 MB. Example 971bd55b-5cbd-43d2-899e-d4a2a7d0a883.
The underlying issue was how it decoded the worksheet XML into large structures. There was no easy fix for that.

* https://github.com/unidoc/unioffice is available as inactive implementation below, although it was found to also use lots of RAM.

* https://github.com/360EntSecGroup-Skylar/excelize was not tested in detail, but seems very similar to "tealeg/xlsx".

* https://github.com/szyhf/go-excel is faster and uses smaller resources than "tealeg/xlsx", but lacks quality when extracting cells and misses many.

*/

package fileconversion

import (
	"bytes"
	"io"

	"github.com/tealeg/xlsx"
)

// IsFileXLSX checks if the data indicates a XLSX file
// XLSX has a signature of 50 4B 03 04
// Warning: This collides with ZIP, DOCX and other zip-based files.
func IsFileXLSX(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04})
}

// XLSX2Text extracts text of an Excel sheet
// Size is the full size of the input file. Limit is the output limit in bytes.
// rowLimit defines how many rows per sheet to extract. -1 means unlimited. This exists as protection against some XLSX files that may use excessive amount of memory.
func XLSX2Text(file io.ReaderAt, size int64, writer io.Writer, limit int64, rowLimit int) (written int64, err error) {
	var xlFile *xlsx.File

	if rowLimit == -1 {
		xlFile, err = xlsx.OpenReaderAt(file, size)
	} else {
		xlFile, err = xlsx.OpenReaderAtWithRowLimit(file, size, rowLimit)
	}
	if err != nil {
		return 0, err
	}

	for n, sheet := range xlFile.Sheets {
		if err = writeOutput(writer, []byte(xlGenerateSheetTitle(sheet.Name, n, int(sheet.MaxRow))), &written, &limit); err != nil || limit == 0 {
			return written, err
		}

		for _, row := range sheet.Rows {

			rowText := ""

			// go through all columns
			for m, cell := range row.Cells {
				text := cell.String()
				if text != "" {
					text = cleanCell(text)

					if m > 0 {
						rowText += ", "
					}
					rowText += text
				}
			}

			rowText += "\n"

			if err = writeOutput(writer, []byte(rowText), &written, &limit); err != nil || limit == 0 {
				return written, err
			}
		}
	}

	return written, nil
}

// XLSX2Cells converts an XLSX file to individual cells
// Size is the full size of the input file.
// rowLimit defines how many rows per sheet to extract. -1 means unlimited. This exists as protection against some XLSX files that may use excessive amount of memory.
func XLSX2Cells(file io.ReaderAt, size int64, rowLimit int) (cells []string, err error) {
	var xlFile *xlsx.File

	if rowLimit == -1 {
		xlFile, err = xlsx.OpenReaderAt(file, size)
	} else {
		xlFile, err = xlsx.OpenReaderAtWithRowLimit(file, size, rowLimit)
	}
	if err != nil {
		return nil, err
	}

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				if text := cell.String(); text != "" {
					text = cleanCell(text)
					cells = append(cells, text)
				}
			}
		}
	}

	return
}

// alternative implementation using https://github.com/unidoc/unioffice, not required

/*
// XLSX2Cells2 converts an XLS file to individual cells
func XLSX2Cells2(file io.ReaderAt, size int64) (cells []string, err error) {

	xlFile, err := spreadsheet.Read(file, size)
	if err != nil || xlFile == nil {
		return nil, err
	}
	defer xlFile.Close()

	for _, sheet := range xlFile.Sheets() {
		for _, row := range sheet.Rows() {
			for _, cell := range row.Cells() {
				if text := cell.GetString(); text != "" {
					text = cleanCell(text)
					cells = append(cells, text)
				}
			}
		}
	}

	return
}

// XLSX2Text2 extracts text from an Excel sheet. It returns bytes written.
// The parameter limit is the max amount of bytes (not characters) to write out.
// The whole Excel file is required even for partial text extraction. This function returns no error with 0 bytes written in case of corrupted or invalid file.
func XLSX2Text2(file io.ReaderAt, size int64, writer io.Writer, limit int64) (written int64, err error) {

	xlFile, err := spreadsheet.Read(file, size)
	if err != nil || xlFile == nil {
		return 0, err
	}
	defer xlFile.Close()

	for n, sheet := range xlFile.Sheets() {
		rows := sheet.Rows()
		if err = writeOutput(writer, []byte(xlGenerateSheetTitle(sheet.Name(), n, len(rows))), &written, &limit); err != nil || limit == 0 {
			return written, err
		}

		for _, row := range sheet.Rows() {
			rowText := ""

			for n, cell := range row.Cells() {
				text := cell.GetString()
				text = cleanCell(text)

				if n > 1 {
					rowText += ", "
				}
				rowText += text
			}

			rowText += "\n"

			if err = writeOutput(writer, []byte(rowText), &written, &size); err != nil || size == 0 {
				return written, err
			}
		}
	}

	return written, nil
}
*/

// implementation using https://github.com/szyhf/go-excel

/*
// XLSX2Text extracts text of an Excel sheet
// Size is the full size of the input file.
func XLSX2Text(file io.ReaderAt, size int64, writer io.Writer, limit int64) (written int64, err error) {

	conn := excel.NewConnecter()
	err = conn.OpenReader(file, size)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	for n, sheetName := range conn.GetSheetNames() {
		if err = writeOutput(writer, []byte(xlGenerateSheetTitle(sheetName, n, 100)), &written, &limit); err != nil || limit == 0 {
			return written, err
		}

		rd, err := conn.NewReader(sheetName)
		if err != nil {
			continue
		}

		for rd.Next() {
			var rowCells []string
			err1 := rd.Read(&rowCells)

			rowText := ""

			// go through all columns
			for m, cell := range rowCells {
				if text := cell; text != "" {
					text = cleanCell(text)

					if m > 0 {
						rowText += ", "
					}
					rowText += text
				}
			}

			rowText += "\n"

			if err = writeOutput(writer, []byte(rowText), &written, &limit); err != nil || limit == 0 {
				return written, err
			}

			if err1 == io.EOF {
				break
			}
		}
	}

	return written, nil
}

// XLSX2Cells converts an XLSX file to individual cells
// Size is the full size of the input file.
func XLSX2Cells(file io.ReaderAt, size int64) (cells []string, err error) {

	conn := excel.NewConnecter()
	err = conn.OpenReader(file, size)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

loopSheet:
	for _, sheetName := range conn.GetSheetNames() {
		rd, err := conn.NewReader(sheetName)
		if err != nil {
			continue
		}

		for rd.Next() {
			var rowCells []string
			if err := rd.Read(&rowCells); err != nil {
				continue loopSheet
			}

			for _, cell := range rowCells {
				if text := cell; text != "" {
					text = cleanCell(text)
					cells = append(cells, text)
				}
			}
		}
	}

	return
}
*/
