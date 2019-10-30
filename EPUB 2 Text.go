/*
File Name:  EPUB 2 Text.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner

EPUB files are ZIP based and contain the content as HTML files.

Tested but did not work:
* https://github.com/n3integration/epub could not read 2 sample files. Also no NewReader function available.

This one was tested and works:
* https://github.com/taylorskalyo/goreader/tree/master/epub

Sample files via https://github.com/IDPF/epub3-samples/releases.
*/

package fileconversion

import (
	"io"

	"github.com/taylorskalyo/goreader/epub"
)

// EPUB2Text converts an EPUB ebook to text
func EPUB2Text(file io.ReaderAt, size int64, limit int64) (string, error) {
	text := ""

	rc, err := epub.NewReader(file, size)
	if err != nil {
		return "", nil
	}

	// The rootfile (content.opf) lists all of the contents of an epub file.
	// There may be multiple rootfiles, although typically there is only one.
	book := rc.Rootfiles[0]

	// Print book title.
	title := "Title: " + book.Title + "\n\n"
	limit -= int64(len(title))
	if limit <= 0 {
		return title, nil
	}

	// List the IDs of files in the book's spine.
	for _, item := range book.Spine.Itemrefs {
		// item.ID was observed to be in one book: cover,titlepage,brief-toc,xpreface_001,xintroduction_001,xepigraph_001,xchapter_001
		reader2, err := item.Open()
		if err != nil {
			continue
		}

		itemText, _ := HTML2Text(reader2)

		// check max length
		if limit <= int64(len(itemText)) {
			itemText = itemText[:limit]
			return title + text, nil
		}

		text += itemText
		limit -= int64(len(itemText))
	}

	if text == "" {
		return "", nil
	}

	return title + text, nil
}
