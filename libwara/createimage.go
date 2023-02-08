package libwara

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/draw"
)

type ImageSize uint8

// TGA header data
type Header struct {
	IdLength     byte
	ColorMapType byte
	ImageType    byte

	ColorMapFirstEntry uint16
	ColorMapLength     uint16
	ColorMapEntrySize  byte

	XOrigin         uint16
	YOrigin         uint16
	ImageWidth      uint16
	ImageHeight     uint16
	PixelDepth      byte
	ImageDescriptor byte
}

// TGA footer data
var Footer = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x54, 0x52, 0x55, 0x45, 0x56, 0x49, 0x53, 0x49, 0x4F, 0x4E, 0x2D, 0x58, 0x46, 0x49, 0x4C, 0x45, 0x2E, 0x00}

// Possible image sizes
const (
	IconSize    ImageSize = iota // Image will be WaraWara Plaza icon
	MessageSize                  // Image will be WaraWara message
	CustomSize                   // Image will be a custom size
)

// Does the bare minimum to create a TGA image
func convertTga(w io.Writer, img image.Image) error {
	header := Header{
		IdLength:           0x00,
		ColorMapType:       0x00,
		ImageType:          0x02,
		ColorMapFirstEntry: 0x0000,
		ColorMapLength:     0x0000,
		ColorMapEntrySize:  0x00,
		XOrigin:            0x0000,
		YOrigin:            0x0000,
		ImageWidth:         uint16(img.Bounds().Dx()),
		ImageHeight:        uint16(img.Bounds().Dy()),
		PixelDepth:         0x20,
		ImageDescriptor:    0x00,
	}

	err := binary.Write(w, binary.LittleEndian, &header)
	if err != nil {
		return err
	}

	startX := 0
	startY := img.Bounds().Dy() - 1
	strideX := 1
	strideY := -1

	for y := startY; 0 <= y && y < img.Bounds().Dy(); y += strideY {
		for x := startX; 0 <= x && x < img.Bounds().Dx(); x += strideX {
			color := img.At(x, y)

			r, g, b, a := color.RGBA()
			colorBytes := []byte{byte(b), byte(g), byte(r), byte(a)}

			err := binary.Write(w, binary.BigEndian, colorBytes)
			if err != nil {
				return err
			}
		}
	}

	err = binary.Write(w, binary.LittleEndian, Footer)
	if err != nil {
		return err
	}

	return nil
}

// Takes a string (input image path) and two int's (output image x and y size), and returns a pointer to a byte buffer and an error
// returns a nil error if no errors occurred
func makeImage(inPath string, xSize, ySize int) (*bytes.Buffer, error) {
	// Open image file
	imageFileIn, err := os.Open(inPath)
	if err != nil {
		return nil, err
	}
	defer imageFileIn.Close()

	// Read image file
	imgData, _, err := image.Decode(imageFileIn)
	if err != nil {
		return nil, err
	}

	// Scale image
	dst := image.NewNRGBA(image.Rect(0, 0, xSize, ySize))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.ApproxBiLinear.Scale(dst, dst.Rect, imgData, imgData.Bounds(), draw.Over, nil)
	imgData = dst

	// Convert image to tga
	var imageFileOut bytes.Buffer
	err = convertTga(&imageFileOut, imgData)
	if err != nil {
		return nil, err
	}

	return &imageFileOut, nil
}

// Converts a byte Buffer to an encoded image
// Returns the encoded image as a String, and an Error
func convertToNintendo(fileData bytes.Buffer) (string, error) {
	var compressedBytes bytes.Buffer
	writer, _ := zlib.NewWriterLevel(&compressedBytes, 6)
	_, err := writer.Write(fileData.Bytes())
	if err != nil {
		return "", err
	}
	writer.Close()

	encodedData := base64.StdEncoding.EncodeToString(compressedBytes.Bytes())

	return encodedData, nil
}

// Encodes an image to the WaraWaraPlaza format
// Takes an image path (String), the type of image (ImageSize), and the dimensions of an image if the image type is CustomImage
// Returns a String of encoded image and an Error
func CreateImage(inPath string, imageSize ImageSize, size ...int) (string, error) {
	xSize, ySize := 0, 0
	switch imageSize {
	case IconSize:
		xSize, ySize = 128, 128
	case MessageSize:
		xSize, ySize = 300, 100
	case CustomSize:
		if len(size) < 2 {
			return "", errors.New("need to supply output image dimensions")
		}
		xSize, ySize = size[0], size[1]
	}

	imgBytes, err := makeImage(inPath, xSize, ySize)
	if err != nil {
		return "", err
	}

	return convertToNintendo(*imgBytes)
}
