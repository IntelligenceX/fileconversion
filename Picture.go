/*
File Name:  Picture.go
Copyright:  2018 Kleissner Investments s.r.o.
Author:     Peter Kleissner
*/

package fileconversion

import (
	"bytes"
	"image"
	_ "image/gif" // automatic registration
	"image/jpeg"
	_ "image/png" // REQUIRED! automatic registration of PNG decoding for image.Decode

	_ "golang.org/x/image/bmp"  // Required for BMP decoding
	_ "golang.org/x/image/tiff" // Required for TIFF decoding

	"github.com/nfnt/resize"
)

// IsExcessiveLargePicture checks if the picture has reasonable width and height, preventing potential DoS when decoding it
// This protects against this problem: If the image claims to be large (in terms of width & height), jpeg.Decode may use a lot of memory, see https://github.com/golang/go/issues/10532.
func IsExcessiveLargePicture(Picture []byte) (excessive bool, err error) {
	config, _, err := image.DecodeConfig(bytes.NewBuffer(Picture))
	if err != nil {
		return false, err
	}

	return config.Width > 7680 || config.Height > 4320, nil
}

// CompressJPEG compresses a JPEG picture according to the input
// Warning: If the image claims to be large (in terms of width & height), this may use a lot of memory. Use IsExcessiveLargePicture first.
func CompressJPEG(Picture []byte, quality int) (compressed []byte) {
	if quality == 100 { // nothing todo on perfect quality
		return Picture
	}

	image, err := jpeg.Decode(bytes.NewBuffer(Picture))
	if err != nil {
		return Picture
	}

	target := bytes.NewBuffer(make([]byte, 0, len(Picture)))

	err = jpeg.Encode(target, image, &jpeg.Options{Quality: quality})
	if err != nil {
		return Picture
	}

	return target.Bytes()
}

// ResizeCompressPicture scales a picture down and compresses it. It accepts GIF, JPEG, PNG as input but output will always be JPEG.
// Quality specifies the output JPEG quality 0-100. Anything below 75 will noticably reduce the picture quality.
// Warning: If the image claims to be large (in terms of width & height), this may use a lot of memory. Use IsExcessiveLargePicture first.
// Scaling a picture down is optional and only done if MaxWidth and MaxHeight are not 0. Even without rescaling, this function is useful to convert a picture into JPEG.
func ResizeCompressPicture(Picture []byte, Quality int, MaxWidth, MaxHeight uint) (compressed []byte, err error) {

	// decode the image
	img, _, err := image.Decode(bytes.NewBuffer(Picture))
	if err != nil { // discard images that can't be decoded
		return nil, err
	}

	// resize if required
	if MaxWidth != 0 && MaxHeight != 0 {
		img = resize.Thumbnail(MaxWidth, MaxHeight, img, resize.Lanczos3)
	}

	// encode as JPEG with the specified quality
	target := bytes.NewBuffer(make([]byte, 0, len(Picture)))

	err = jpeg.Encode(target, img, &jpeg.Options{Quality: Quality})
	if err != nil {
		return nil, err
	}

	return target.Bytes(), nil
}
