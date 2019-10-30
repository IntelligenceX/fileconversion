/*
File Name:  ODT 2 Text.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner

Fork from https://github.com/lu4p/cat/blob/master/odtxt/odtreader.go.
The extract discards any formatting. The output is one large string without new-lines at the current time.
*/

package fileconversion

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"

	"github.com/IntelligenceX/fileconversion/html2text"
)

// ODT2Text extracts text of an OpenDocument Text file
// Size is the full size of the input file.
func ODT2Text(file io.ReaderAt, size int64, writer io.Writer, limit int64) (written int64, err error) {
	f, err := odtNewReader(file, size)
	if err != nil {
		return 0, err
	}

	text, err := f.GetTxt()
	if err != nil {
		return 0, err
	}

	err = writeOutput(writer, []byte(text), &written, &limit)

	return
}

//odt zip struct
type odt struct {
	zipFileReader *zip.Reader
	Files         []*zip.File
	FilesContent  map[string][]byte
	Content       string
}

func odtNewReader(file io.ReaderAt, size int64) (*odt, error) {
	reader, err := zip.NewReader(file, size)
	if err != nil {
		return nil, err
	}

	odtDoc := odt{
		zipFileReader: reader,
		Files:         reader.File,
		FilesContent:  map[string][]byte{},
	}

	for _, f := range odtDoc.Files {
		contents, _ := odtDoc.retrieveFileContents(f.Name)
		odtDoc.FilesContent[f.Name] = contents
	}

	return &odtDoc, nil
}

//Read all files contents
func (d *odt) retrieveFileContents(filename string) ([]byte, error) {
	var file *zip.File
	for _, f := range d.Files {
		if f.Name == filename {
			file = f
			break
		}
	}

	if file == nil {
		return nil, errors.New(filename + " file not found")
	}

	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func (d *odt) GetTxt() (content string, err error) {
	xmlData := d.FilesContent["content.xml"]
	return xml2Text(xmlData)
	//content, err = d.listP(xmlData)
}

/*
// listP for w:p tag value
func (d *odt) listP(data []byte) (string, error) {
	v := new(odtQuery)
	err := xml.Unmarshal(data, &v)
	if err != nil {
		return "", err
	}
	var result string
	for _, text := range v.Body.Text {
		for _, line := range text.P {
			if line == "" {
				continue
			}
			result += line + "\n"
		}
	}
	return result, nil
}

type odtQuery struct {
	XMLName xml.Name `xml:"document-content"`
	Body    odtBody  `xml:"body"`
}
type odtBody struct {
	Text []odtText `xml:"text"`
}
type odtText struct {
	P []string `xml:"p"`
}
*/

// xml2Text extracts any text from XML data.
// Note that any formatting will be lost. The output is one large string without new-lines.
func xml2Text(data []byte) (string, error) {
	return html2text.FromString(string(data))
}
