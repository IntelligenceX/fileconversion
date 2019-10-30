/*
File Name:  MOBI 2 Text.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner

Mobi files use HTML tags.

Did not work:
* https://github.com/766b/mobi is only a writer and does not have a useful reader
* https://github.com/peterbn/mobi a fork of above one.

Works:
* https://github.com/neofight/mobi code basically works, just an in-memory open function had to be forked.

*/

package fileconversion

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	html "github.com/levigross/exp-html"
	"github.com/neofight/mobi/convert"
	"github.com/neofight/mobi/headers"
)

// Mobi2Text converts a MOBI ebook to text
func Mobi2Text(file io.ReadSeeker) (string, error) {

	book, _ := mobiOpen(file)
	markupText, _ := book.Markup()

	text, _ := HTML2Text(strings.NewReader(markupText))

	return text, nil
}

// below code is forked from https://github.com/neofight/mobi MOBIFile.go

type mobiBook struct {
	file          io.ReadSeeker
	pdbHeader     *headers.PDB
	palmDOCHeader *headers.PalmDOC
	mobiHeader    *headers.MOBI
	exthHeader    *headers.EXTH
}

func mobiOpen(file io.ReadSeeker) (*mobiBook, error) {

	var book mobiBook

	var err error

	book.file = file
	book.pdbHeader, err = headers.ReadPDB(book.file)

	if err != nil {
		return nil, fmt.Errorf("unable to read PDB header: %v", err)
	}

	book.palmDOCHeader, err = headers.ReadPalmDOC(book.file)

	if err != nil {
		return nil, fmt.Errorf("unable to read PalmDOC header: %v", err)
	}

	book.mobiHeader, err = headers.ReadMOBI(book.file)

	if err != nil {
		return nil, fmt.Errorf("unable to read MOBI header: %v", err)
	}

	if book.mobiHeader.EXTHHeaderPresent {

		book.exthHeader, err = headers.ReadEXTH(book.file)

		if err != nil {
			return nil, fmt.Errorf("unable to read EXTH header: %v", err)
		}
	}

	return &book, nil
}

func (mobiFile mobiBook) Cover() ([]byte, error) {

	for _, r := range mobiFile.exthHeader.Records {

		if r.RecordType == 201 {
			coverIndex := mobiFile.mobiHeader.FirstImageIndex + convert.FromUint32(r.RecordData)

			record := mobiFile.pdbHeader.Records[coverIndex]
			nextRecord := mobiFile.pdbHeader.Records[coverIndex+1]

			coverOffset := record.RecordDataOffset
			coverSize := nextRecord.RecordDataOffset - coverOffset

			_, err := mobiFile.file.Seek(int64(coverOffset), 0)

			if err != nil {
				return nil, fmt.Errorf("unable to find cover: %v", err)
			}

			cover := make([]byte, coverSize)

			err = binary.Read(mobiFile.file, binary.BigEndian, &cover)

			if err != nil {
				return nil, fmt.Errorf("unable to read cover: %v", err)
			}

			return cover, nil
		}
	}

	return nil, nil
}

func (mobiFile mobiBook) Markup() (string, error) {

	startIndex := mobiFile.mobiHeader.FirstContentIndex
	endIndex := mobiFile.mobiHeader.FirstNonBookIndex - 1

	if endIndex > len(mobiFile.pdbHeader.Records)-2 {
		endIndex = len(mobiFile.pdbHeader.Records) - 2
	}

	if endIndex < 0 || startIndex < 0 || startIndex >= len(mobiFile.pdbHeader.Records) {
		return "", fmt.Errorf("Invalid header")
	}

	text := make([]byte, 0)

	for index := startIndex; index <= endIndex; index++ {

		record := mobiFile.pdbHeader.Records[index]
		nextRecord := mobiFile.pdbHeader.Records[index+1]

		recordOffset := record.RecordDataOffset
		recordSize := nextRecord.RecordDataOffset - recordOffset

		_, err := mobiFile.file.Seek(int64(recordOffset), 0)

		if err != nil {
			return "", fmt.Errorf("unable to find text: %v", err)
		}

		recordData := make([]byte, recordSize)

		err = binary.Read(mobiFile.file, binary.BigEndian, &recordData)

		if err != nil {
			return "", fmt.Errorf("unable to read text: %v", err)
		}

		recordText := fromLZ77(recordData)

		text = append(text, recordText...)
	}

	text = text[:mobiFile.palmDOCHeader.TextLength]

	if !utf8.Valid(text) {
		return "", errors.New("unable to decompress text")
	}

	return string(text), nil
}

func (mobiFile mobiBook) Text() (string, error) {

	markup, err := mobiFile.Markup()

	if err != nil {
		return "", fmt.Errorf("unable to read markup: %v", err)
	}

	pos, err := getTOCPosition(markup)

	if err != nil {
		return "", fmt.Errorf("unable to locate TOC: %v", err)
	}

	bookmarks, err := parseTOC(markup[pos:])

	text := make([]string, 0)

	for i := range bookmarks {

		start := bookmarks[i]
		var end int

		if i < len(bookmarks)-1 {
			end = bookmarks[i+1]
		} else {
			end = pos
		}

		paragraphs, err := parseChapter(markup[start:end])

		if err != nil {
			return "", fmt.Errorf("unable to parse chapter: %v", err)
		}

		text = append(text, paragraphs...)
	}

	return strings.Join(text, "\n\n"), nil
}

func getTOCPosition(markup string) (int, error) {

	htmlReader := strings.NewReader(markup)

	tokenizer := html.NewTokenizer(htmlReader)

	for {
		tokenType := tokenizer.Next()

		switch {
		case tokenType == html.ErrorToken:
			return 0, fmt.Errorf("unable to find reference element")
		case tokenType == html.SelfClosingTagToken:
			token := tokenizer.Token()

			if token.Data == "reference" {
				filepos, err := attr(token, "filepos")

				if err != nil {
					return 0, errors.New("filepos attribute missing")
				}

				pos, err := strconv.Atoi(filepos)

				if err != nil {
					return 0, errors.New("filepos attribute invalid")
				}

				return pos, nil
			}
		}
	}
}

func parseTOC(markup string) ([]int, error) {

	toc := make([]int, 0)

	htmlReader := strings.NewReader(markup)

	tokenizer := html.NewTokenizer(htmlReader)

	for {
		tokenType := tokenizer.Next()

		switch {
		case tokenType == html.ErrorToken:
			return toc[1:], nil
		case tokenType == html.StartTagToken:
			token := tokenizer.Token()

			if token.Data == "a" {
				filepos, err := attr(token, "filepos")

				if err != nil {
					continue
				}

				pos, err := strconv.Atoi(filepos)

				if err != nil {
					return nil, errors.New("filepos attribute invalid")
				}

				toc = append(toc, pos)
			}
		}
	}
}

func parseChapter(markup string) ([]string, error) {

	paragraphs := make([]string, 0)

	htmlReader := strings.NewReader(markup)

	tokenizer := html.NewTokenizer(htmlReader)

	for {
		tokenType := tokenizer.Next()

		switch {
		case tokenType == html.ErrorToken:
			return paragraphs, nil
		case tokenType == html.TextToken:
			token := tokenizer.Token()

			if len(strings.TrimSpace(token.Data)) > 0 {
				paragraphs = append(paragraphs, strings.TrimSpace(token.Data))
			}
		}
	}
}

func attr(t html.Token, name string) (string, error) {
	for _, a := range t.Attr {
		if a.Key == name {
			return a.Val, nil
		}
	}

	return "", fmt.Errorf("attribute %v not found", name)
}

// fromLZ77 is forked from conversion.go because of index out of range panic
func fromLZ77(text []byte) []byte {

	var reader = bytes.NewReader(text)

	var buffer [4096]byte
	var pos int

	for {
		if pos == 4096 {
			break
		}

		c, err := reader.ReadByte()

		if err == io.EOF {
			break
		}

		switch {

		// 0x00: "1 literal" copy that byte unmodified to the decompressed stream.
		case c == 0x00:
			buffer[pos] = c
			pos++

		// 0x09 to 0x7f: "1 literal" copy that byte unmodified to the decompressed stream.
		case c >= 0x09 && c <= 0x7f:
			buffer[pos] = c
			pos++

		// 0x01 to 0x08: "literals": the byte is interpreted as a count from 1 to 8, and that many literals are copied
		// unmodified from the compressed stream to the decompressed stream.
		case c >= 0x01 && c <= 0x08:
			length := int(c)
			for i := 0; i < length; i++ {
				c, err = reader.ReadByte()
				buffer[pos] = c
				pos++
			}

		// 0x80 to 0xbf: "length, distance" pair: the 2 leftmost bits of this byte ('10') are discarded, and the
		// following 6 bits are combined with the 8 bits of the next byte to make a 14 bit "distance, length" item.
		// Those 14 bits are broken into 11 bits of distance backwards from the current location in the uncompressed
		// text, and 3 bits of length to copy from that point (copying n+3 bytes, 3 to 10 bytes).
		case c >= 0x80 && c <= 0xbf:
			c2, _ := reader.ReadByte()

			distance := (int(c&0x3F)<<8 | int(c2)) >> 3
			length := int(c2&0x07) + 3

			start := pos - distance

			for i := 0; i < length; i++ {
				// check if index is in range
				if start+i >= len(buffer) || start+i < 0 {
					return buffer[:pos]
				}

				c = buffer[start+i]
				buffer[pos] = c
				pos++
			}

		// 0xc0 to 0xff: "byte pair": this byte is decoded into 2 characters: a space character, and a letter formed
		// from this byte XORed with 0x80.
		case c >= 0xc0:
			buffer[pos] = ' '
			pos++
			buffer[pos] = c ^ 0x80
			pos++
		}
	}

	return buffer[:pos]
}

// IsFileMOBI checks if the data indicates a MOBI file
func IsFileMOBI(data []byte) bool {
	// Mobi files have a header and there is the signature "BOOKMOBI" or "TEXtREAd".
	// There are many more more potential signatures https://sno.phy.queensu.ca/~phil/exiftool/TagNames/Palm.html

	// Fork from code here http://will.tip.dhappy.org/lib/calibre/dedrm/mobidedrm.py
	// if self.header[0x3C:0x3C+8] != 'BOOKMOBI' and self.header[0x3C:0x3C+8] != 'TEXtREAd':
	//   raise DrmException(u"Invalid file format")

	if len(data) < 0x3C+8 {
		return false
	}

	signature := data[0x3C : 0x3C+8]

	return bytes.Equal(signature, []byte("BOOKMOBI")) || bytes.Equal(signature, []byte("TEXtREAd"))
}
