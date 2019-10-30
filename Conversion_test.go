/*
File Name:  Conversion_test.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner
*/

package fileconversion

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestXLS(t *testing.T) {
	// open local file to extract text and output to command line
	file, err := os.Open("test.xls")
	if err != nil {
		return
	}

	defer file.Close()

	XLS2Text(file, os.Stdout, 1*1024*1024)
}

func TestPPTX(t *testing.T) {
	// open local file to extract text and output to command line
	file, err := os.Open("test.pptx")
	if err != nil {
		return
	}

	defer file.Close()

	stat, _ := file.Stat()

	text, _ := PPTX2Text(file, stat.Size())
	fmt.Print(text)
}

func TestODS(t *testing.T) {
	// open local file to extract text and output to command line
	file, err := os.Open("test.ods")
	if err != nil {
		return
	}

	defer file.Close()
	stat, _ := file.Stat()

	ODS2Text(file, stat.Size(), os.Stdout, 1*1024*1024)
}

func TestExcelCell(t *testing.T) {
	file1, err := os.Open("test.xls")
	if err == nil {
		cells, _ := XLS2Cells(file1)
		for n, cell := range cells {
			fmt.Printf("%s\n", cell)
			if n > 20 {
				break
			}
		}

		file1.Close()
	}

	file1, err = os.Open("test.xlsx")
	if err == nil {
		stat, _ := file1.Stat()
		cells, _ := XLSX2Cells(file1, stat.Size(), 1000)
		for n, cell := range cells {
			fmt.Printf("%s\n", cell)
			if n > 20 {
				break
			}
		}

		file1.Close()
	}

	file1, err = os.Open("test.ods")
	if err == nil {
		stat, _ := file1.Stat()
		cells, _ := ODS2Cells(file1, stat.Size())
		for n, cell := range cells {
			fmt.Printf("%s\n", cell)
			if n > 20 {
				break
			}
		}

		file1.Close()
	}

}

func TestCSV(t *testing.T) {
	file, err := os.Open("test.txt")
	if err != nil {
		return
	}
	defer file.Close()

	content, _ := ioutil.ReadAll(file)

	IsCSV(content)
}

func TestEPUB(t *testing.T) {
	// open local file to extract text and output to command line
	file, err := os.Open("moby-dick.epub")
	if err != nil {
		return
	}

	defer file.Close()

	stat, _ := file.Stat()

	text, _ := EPUB2Text(file, stat.Size(), 1000)
	fmt.Print(text)
}

func TestMOBI(t *testing.T) {
	// open local file to extract text and output to command line
	file, err := os.Open("windows-1252.mobi")
	if err != nil {
		return
	}

	defer file.Close()

	text, _ := Mobi2Text(file)
	fmt.Print(text)
}

func TestPDFImage(t *testing.T) {
	// open local file to extract images
	file, err := os.Open("test.pdf")
	if err != nil {
		return
	}

	defer file.Close()

	images, _ := PDFExtractImages(file)
	fmt.Print(len(images))
}

func TestPD2Text(t *testing.T) {
	file, err := os.Open("1.pdf")
	if err != nil {
		return
	}

	defer file.Close()

	buffer := bytes.NewBuffer(make([]byte, 0, 2*1024))
	PDFListContentStreams(file, buffer, 2*1024)

	fmt.Println(buffer.String())
}

func TestODTText(t *testing.T) {
	file, err := os.Open("Test\\file-sample_500kB.odt")
	if err != nil {
		return
	}

	defer file.Close()
	stat, _ := file.Stat()

	buffer := bytes.NewBuffer(make([]byte, 0, 2*1024))

	ODT2Text(file, stat.Size(), buffer, 2*1024)

	fmt.Println(buffer.String())
}

// TestXLSX extracts text from an XLSX file.
// Memory usage: 100 rows = 52 MB, 500 rows = 200 MB, 1000 rows = 400 MB, 2000/5000/10000/-1 rows = 700 MB
func TestXLSX(t *testing.T) {
	file, err := os.Open("Test\\971bd55b-5cbd-43d2-899e-d4a2a7d0a883.xlsx")
	if err != nil {
		return
	}

	defer file.Close()
	stat, _ := file.Stat()

	buffer := bytes.NewBuffer(make([]byte, 0, 2*1024))

	XLSX2Text(file, stat.Size(), buffer, 2*1024, -1)

	fmt.Println(buffer.String())
}
