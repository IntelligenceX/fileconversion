# ole2
Microsoft Compound Document File Format library in Golang

Forked from https://github.com/extrame/ole2.

The ole code had major bugs. Slice bounds were not checked and caused crashes.

The code was adapted to resemble https://github.com/ElevenPaths/FOCA/blob/master/MetadataExtractCore/Metadata/OleDocument.cs.
Alternative implementation for reference: https://github.com/sassoftware/relic/blob/4db78dcc59ae33d7565f3927e4c7bc8a86ee146c/lib/comdoc/msat.go
