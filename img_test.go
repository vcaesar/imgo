package imgo

import (
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
