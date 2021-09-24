package imgo

import (
	"errors"
	"image"
)

// EncodeImg encode the image.Image return pix and stride
func EncodeImg(m image.Image) (pix []uint8, stride int, err error) {
	d := m.Bounds().Size()
	if d.X < 0 || d.Y < 0 {
		err = errors.New("imgo: negative bounds")
		return
	}
	h := &header{
		sigBM:         [2]byte{'B', 'M'},
		fileSize:      14 + 40,
		pixOffset:     14 + 40,
		dibHeaderSize: 40,
		width:         uint32(d.X),
		height:        uint32(d.Y),
		colorPlane:    1,
	}

	var step int
	var palette []byte
	var opaque bool
	switch m := m.(type) {
	case *image.Gray:
		step = (d.X + 3) &^ 3
		palette = make([]byte, 1024)
		for i := 0; i < 256; i++ {
			palette[i*4+0] = uint8(i)
			palette[i*4+1] = uint8(i)
			palette[i*4+2] = uint8(i)
			palette[i*4+3] = 0xFF
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += uint32(len(palette)) + h.imageSize
		h.pixOffset += uint32(len(palette))
		h.bpp = 8

	case *image.Paletted:
		step = (d.X + 3) &^ 3
		palette = make([]byte, 1024)
		for i := 0; i < len(m.Palette) && i < 256; i++ {
			r, g, b, _ := m.Palette[i].RGBA()
			palette[i*4+0] = uint8(b >> 8)
			palette[i*4+1] = uint8(g >> 8)
			palette[i*4+2] = uint8(r >> 8)
			palette[i*4+3] = 0xFF
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += uint32(len(palette)) + h.imageSize
		h.pixOffset += uint32(len(palette))
		h.bpp = 8
	case *image.RGBA:
		opaque = m.Opaque()
		if opaque {
			step = (3*d.X + 3) &^ 3
			h.bpp = 24
		} else {
			step = 4 * d.X
			h.bpp = 32
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += h.imageSize
	case *image.NRGBA:
		opaque = m.Opaque()
		if opaque {
			step = (3*d.X + 3) &^ 3
			h.bpp = 24
		} else {
			step = 4 * d.X
			h.bpp = 32
		}
		h.imageSize = uint32(d.Y * step)
		h.fileSize += h.imageSize
	default:
		step = (3*d.X + 3) &^ 3
		h.imageSize = uint32(d.Y * step)
		h.fileSize += h.imageSize
		h.bpp = 24
	}

	if d.X == 0 || d.Y == 0 {
		return
	}

	switch m := m.(type) {
	case *image.Gray:
		stride = m.Stride
		pix = m.Pix
	case *image.Paletted:
		stride = m.Stride
		pix = m.Pix
	case *image.RGBA:
		stride = m.Stride
		pix = m.Pix
	case *image.NRGBA:
		stride = m.Stride
		pix = m.Pix
	default:
		// stride = m.Stride
		// pix = m.Pix
	}

	return
}

// ConvertToRGBA convert the image.Image to *image.RGBA
func ConvertToRGBA(img image.Image) (r *image.RGBA) {
	pix, stride, _ := EncodeImg(img)
	return &image.RGBA{
		Pix:    pix,
		Stride: stride,
		Rect:   image.Rect(0, 0, Width(img), Height(img)),
	}
}

type header struct {
	sigBM           [2]byte
	fileSize        uint32
	resverved       [2]uint16
	pixOffset       uint32
	dibHeaderSize   uint32
	width           uint32
	height          uint32
	colorPlane      uint16
	bpp             uint16
	compression     uint32
	imageSize       uint32
	xPixelsPerMeter uint32
	yPixelsPerMeter uint32
	colorUse        uint32
	colorImportant  uint32
}
