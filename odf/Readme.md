This code was forked from https://github.com/knieriem/odf.

This projekt contains two Go packages – odf and odf/ods
– that allow basic read-only access to the tables of Open
Document Spreadsheets, making use of Go's encoding/xml package.

For now the ods package makes it easy to convert a table to a
`[][]string`.
