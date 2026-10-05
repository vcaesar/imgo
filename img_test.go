package imgo

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/vcaesar/tt"
)

func TestImg(t *testing.T) {
	img, err := Read("testdata/test_007.jpeg")
	tt.Nil(t, err)

	err = SaveToJpeg("testdata/test_1.jpeg", img)
	tt.Nil(t, err)
	err = Save("testdata/test_1.bmp", img)
	tt.Nil(t, err)
}

// testRGBA returns an opaque w x h image with distinct pixels.
func testRGBA(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: uint8(10 * x), G: uint8(20 * y), B: uint8(x + y), A: 255})
		}
	}
	return img
}

// DecodeFile and ImgToBytes must close the file: Windows refuses to
// delete an open file.
func TestDecodeFileReleasesFile(t *testing.T) {
	dir := t.TempDir()
	for _, open := range []struct {
		name string
		fn   func(string) error
	}{
		{"decode.png", func(p string) error { _, _, err := DecodeFile(p); return err }},
		{"bytes.png", func(p string) error { _, err := ImgToBytes(p); return err }},
	} {
		path := filepath.Join(dir, open.name)
		tt.Nil(t, SaveToPNG(path, testRGBA(2, 2)))
		tt.Nil(t, open.fn(path))
		if err := os.Remove(path); err != nil {
			t.Fatalf("%s left file open: %v", open.name, err)
		}
	}

	bad := filepath.Join(dir, "bad.png")
	tt.Nil(t, os.WriteFile(bad, []byte("not an image"), 0o600))
	_, _, err := DecodeFile(bad)
	tt.NotNil(t, err)
	if err := os.Remove(bad); err != nil {
		t.Fatalf("DecodeFile error path left file open: %v", err)
	}

	_, _, err = DecodeFile(filepath.Join(dir, "missing.png"))
	tt.True(t, errors.Is(err, fs.ErrNotExist))
}

// ToByte returns strict base64 without trailing zero padding.
func TestToByteStrictBase64(t *testing.T) {
	src := testRGBA(3, 2)

	b := ToByte(src, "png")
	raw, err := base64.StdEncoding.Strict().DecodeString(string(b))
	tt.Nil(t, err)
	tt.True(t, bytes.HasPrefix(raw, []byte("\x89PNG")))
	tt.Equal(t, string(b), ToString(src, "png"))

	img, err := StrToImg(string(b))
	tt.Nil(t, err)
	tt.Equal(t, src.Bounds(), img.Bounds())

	raw, err = base64.StdEncoding.Strict().DecodeString(string(ToByte(src)))
	tt.Nil(t, err)
	_, fm, err := image.Decode(bytes.NewReader(raw))
	tt.Nil(t, err)
	tt.Equal(t, "jpeg", fm)

	if ToByte(src, "webp") != nil {
		t.Error("ToByte unsupported format: want nil")
	}
}

func TestGetFm(t *testing.T) {
	tests := map[string]string{
		"a.png":            "png",
		"a.JPG":            "jpg",
		"dir.v1/a.jpeg":    "jpeg",
		"dir.v1/noext":     "",
		"x/y/z.tar.gz.bmp": "bmp",
	}
	for path, want := range tests {
		tt.Equal(t, want, getFm(path))
	}
}

// .jpg, .tif and upper-case extensions save and read back.
func TestSaveReadExtensions(t *testing.T) {
	dir := t.TempDir()
	src := testRGBA(8, 6)
	for _, name := range []string{"a.jpg", "b.JPEG", "c.tif", "d.PNG", "e.bmp", "f.gif"} {
		path := filepath.Join(dir, name)
		if err := Save(path, src, 90); err != nil {
			t.Fatalf("Save %s: %v", name, err)
		}
		img, err := Read(path)
		if err != nil {
			t.Fatalf("Read %s: %v", name, err)
		}
		tt.Equal(t, src.Bounds(), img.Bounds())
	}

	tt.NotNil(t, Save(filepath.Join(dir, "x.webp"), src))
	_, err := Read(filepath.Join(dir, "a.jpg"+".unknown"))
	tt.NotNil(t, err)
}

// Create returns a usable, open file.
func TestCreateReturnsOpenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.txt")
	f, err := Create(path)
	tt.Nil(t, err)
	_, err = f.WriteString("ok")
	tt.Nil(t, err)
	tt.Nil(t, f.Close())

	b, err := os.ReadFile(path)
	tt.Nil(t, err)
	tt.Equal(t, "ok", string(b))
}

func TestBytesRoundTrip(t *testing.T) {
	src := testRGBA(4, 4)
	b, err := ToBytesPng(src)
	tt.Nil(t, err)
	img, err := ByteToImg(b)
	tt.Nil(t, err)
	tt.Equal(t, src.Bounds(), img.Bounds())

	_, err = ToBytes(src, "webp")
	tt.NotNil(t, err)

	path := filepath.Join(t.TempDir(), "p.png")
	tt.Nil(t, SaveByte(path, b))
	got, err := PngToBytes(path)
	tt.Nil(t, err)
	tt.Equal(t, b, got)

	enc, err := OpenBase64(path)
	tt.Nil(t, err)
	out := filepath.Join(t.TempDir(), "q.png")
	tt.Nil(t, SaveByBase64(enc, out))
	got, err = os.ReadFile(out)
	tt.Nil(t, err)
	tt.Equal(t, b, got)
}
