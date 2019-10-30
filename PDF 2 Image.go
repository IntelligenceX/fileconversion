/*
File Name:  PDF 2 Image.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner
*/

package fileconversion

import (
	"image"
	"io"
	"strconv"

	pdfcontent "github.com/unidoc/unipdf/contentstream"
	pdfcore "github.com/unidoc/unipdf/core"
	pdf "github.com/unidoc/unipdf/model"
)

var xObjectImages = 0
var inlineImages = 0

// ImageResult contains an extracted image
type ImageResult struct {
	Image image.Image
	Name  string
}

// PDFExtractImages extracts all images from a PDF file
func PDFExtractImages(input io.ReadSeeker) (images []ImageResult, err error) {

	pdfReader, err := pdf.NewPdfReader(input)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}

	// Try decrypting with an empty one.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			// Encrypted and we cannot do anything about it.
			return nil, err
		}
		if !auth {
			//fmt.Println("Need to decrypt with password")
			return nil, nil
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("PDF Num Pages: %d\n", numPages)

	for i := 0; i < numPages; i++ {
		//fmt.Printf("-----\nPage %d:\n", i+1)

		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return nil, err
		}

		// List images on the page.
		rgbImages, err := extractImagesOnPage(page)
		if err != nil {
			return nil, err
		}
		_ = rgbImages

		for idx, img := range rgbImages {
			fname := "p" + strconv.Itoa(i+1) + "_" + strconv.Itoa(idx) + ".jpg"

			gimg, err := img.ToGoImage()
			if err != nil {
				return nil, err
			}

			images = append(images, ImageResult{Image: gimg, Name: fname})
		}
	}

	return images, nil
}

func extractImagesOnPage(page *pdf.PdfPage) ([]*pdf.Image, error) {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return nil, err
	}

	return extractImagesInContentStream(contents, page.Resources)
}

func extractImagesInContentStream(contents string, resources *pdf.PdfPageResources) ([]*pdf.Image, error) {
	rgbImages := []*pdf.Image{}
	cstreamParser := pdfcontent.NewContentStreamParser(contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		return nil, err
	}

	processedXObjects := map[string]bool{}

	// Range through all the content stream operations.
	for _, op := range *operations {
		if op.Operand == "BI" && len(op.Params) == 1 {
			// BI: Inline image.

			iimg, ok := op.Params[0].(*pdfcontent.ContentStreamInlineImage)
			if !ok {
				continue
			}

			img, err := iimg.ToImage(resources)
			if err != nil {
				return nil, err
			}

			cs, err := iimg.GetColorSpace(resources)
			if err != nil {
				return nil, err
			}
			if cs == nil {
				// Default if not specified?
				cs = pdf.NewPdfColorspaceDeviceGray()
			}
			//fmt.Printf("Cs: %T\n", cs)

			rgbImg, err := cs.ImageToRGB(*img)
			if err != nil {
				return nil, err
			}

			rgbImages = append(rgbImages, &rgbImg)
			inlineImages++
		} else if op.Operand == "Do" && len(op.Params) == 1 {
			// Do: XObject.
			name := op.Params[0].(*pdfcore.PdfObjectName)

			// Only process each one once.
			_, has := processedXObjects[string(*name)]
			if has {
				continue
			}
			processedXObjects[string(*name)] = true

			_, xtype := resources.GetXObjectByName(*name)
			if xtype == pdf.XObjectTypeImage {
				//fmt.Printf(" XObject Image: %s\n", *name)

				ximg, err := resources.GetXObjectImageByName(*name)
				if err != nil {
					return nil, err
				}

				img, err := ximg.ToImage()
				if err != nil {
					return nil, err
				}

				rgbImg, err := ximg.ColorSpace.ImageToRGB(*img)
				if err != nil {
					return nil, err
				}
				rgbImages = append(rgbImages, &rgbImg)
				xObjectImages++
			} else if xtype == pdf.XObjectTypeForm {
				// Go through the XObject Form content stream.
				xform, err := resources.GetXObjectFormByName(*name)
				if err != nil {
					return nil, err
				}

				formContent, err := xform.GetContentStream()
				if err != nil {
					return nil, err
				}

				// Process the content stream in the Form object too:
				formResources := xform.Resources
				if formResources == nil {
					formResources = resources
				}

				// Process the content stream in the Form object too:
				formRgbImages, err := extractImagesInContentStream(string(formContent), formResources)
				if err != nil {
					return nil, err
				}
				rgbImages = append(rgbImages, formRgbImages...)
			}
		}
	}

	return rgbImages, nil
}
