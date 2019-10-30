package odf

import (
	"fmt"
	"os"
	"testing"
)

func ExampleMetaTitle() {
	f, err := Open("./ods/test.ods")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer f.Close()

	m, err := f.Meta()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	t, _ := m.Meta.CreationDate.Time()
	fmt.Println(m.Meta.Title, t.Format("(2006-01-02)"))

	// Output: Test Spreadsheet for odf/ods package (2012-01-10)
}

func TestDummy(_ *testing.T) {
}
