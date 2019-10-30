/*
File Name:  PPTX 2 Text.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner

This code is a fork from https://github.com/mr-tim/rol-o-decks/blob/master/indexer/indexer.go.
*/

package fileconversion

import (
	"archive/zip"
	"bytes"
	"io"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/xmlpath.v2"
)

// PPTXDocument is a PPTX document loaded into memory
type PPTXDocument struct {
	Slides []PPTXSlide
}

// PPTXSlide is a single PPTX slide
type PPTXSlide struct {
	SlideNumber int
	//ThumbnailBase64 string
	TextContent string
}

// SlideNumberSorter is used for sorting
type SlideNumberSorter []PPTXSlide

func (a SlideNumberSorter) Len() int           { return len(a) }
func (a SlideNumberSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SlideNumberSorter) Less(i, j int) bool { return a[i].SlideNumber < a[j].SlideNumber }

// PPTX2Text extracts text of a PowerPoint document
// Size is the full size of the input file.
func PPTX2Text(file io.ReaderAt, size int64) (string, error) {

	r, err := zip.NewReader(file, size)
	if err != nil {
		return "", err
	}

	doc := parsePPTXDocument(r)

	return doc.AsText(), nil
}

// IsFilePPTX checks if the data indicates a PPTX file
// PPTX has a signature of 50 4B 03 04
// Warning: This collides with ZIP, DOCX and other zip-based files.
func IsFilePPTX(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04})
}

func extractSlideContent(f *zip.File) string {
	p := xmlpath.MustCompile("//t")
	zr, _ := f.Open()
	defer zr.Close()
	root, _ := xmlpath.Parse(zr)
	i := p.Iter(root)
	content := make([]string, 0)
	for i.Next() {
		n := i.Node()
		content = append(content, n.String())
	}
	textContent := strings.Join(content, "\n")
	return textContent
}

func parsePPTXDocument(r *zip.Reader) (doc PPTXDocument) {

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/") && !strings.HasPrefix(f.Name, "ppt/slides/_rels") {
			slideNumberStr := strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(f.Name), "ppt/slides/slide"), ".xml")
			slideNumber, _ := strconv.Atoi(slideNumberStr)

			// grab the text content
			doc.Slides = append(doc.Slides, PPTXSlide{
				SlideNumber: slideNumber,
				TextContent: extractSlideContent(f),
				//ThumbnailBase64: generateThumbnail(fileToIndex, slideNumber),
			})
		}
	}

	sort.Sort(SlideNumberSorter(doc.Slides))

	return doc
}

// AsText returns the text on all slides
func (doc PPTXDocument) AsText() (text string) {

	for n, slide := range doc.Slides {
		if slide.TextContent == "" { // skip empty slides
			continue
		}

		if n > 0 && text != "" {
			text += "\n\n"
		}

		text += "Slide " + strconv.Itoa(n+1) + ":\n"
		text += slide.TextContent
	}

	return text
}
