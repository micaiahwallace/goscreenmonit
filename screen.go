package goscreenmonit

import (
	"bytes"
	"compress/zlib"
	"image"
	"io/ioutil"

	"github.com/kbinani/screenshot"
)

// Get a list of capturers for all displays available
// func GetAllCapturers() ([]*scrap.Capturer, error) {

// 	// Get all displays
// 	disps, err := scrap.Displays()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// store capturers
// 	caps := make([]*scrap.Capturer, len(disps))

// 	// Loop over displays
// 	for i, _ := range disps {

// 		// Create capturer for it
// 		d, derr := scrap.GetDisplay(i)
// 		if derr != nil {
// 			log.Println("Unable to get display.", i)
// 			continue
// 		}
// 		cap, err := scrap.NewCapturer(d)
// 		if err != nil {
// 			return nil, err
// 		}
// 		caps[i] = cap
// 	}

// 	return caps, nil
// }

// Capture a screenshot for a screen capturer
// func CaptureScreen_old(c *scrap.Capturer) ([]byte, error) {

// 	// Get an image, trying until one available
// 	for {
// 		if img, _, err := c.FrameImage(); img != nil || err != nil {
// 			// Detach the image so it's safe to use after this method
// 			if img != nil {
// 				img.Detach()
// 			}

// 			// Copy image bytes into a new []byte
// 			imgcopy := make([]byte, len(img.Pix))
// 			copy(imgcopy, img.Pix)
// 			return imgcopy, err
// 		}

// 		// Sleep 17ms (~1/60th of a second) and try again
// 		time.Sleep(17 * time.Millisecond)
// 	}
// }

// Get number of screens
func GetScreenCount() int {
	return screenshot.NumActiveDisplays()
}

// Capture screen by index
func CaptureScreen(index int) (*image.RGBA, error) {

	// Capture display by index
	img, err := screenshot.CaptureDisplay(index)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// Compress screenshot with zlib
func EncodeImage(img []byte) ([]byte, error) {

	// Create buffer and encode image
	var byt bytes.Buffer
	w := zlib.NewWriter(&byt)

	// Write image to encoder
	// log.Printf("Writing image: %v\n", len(img))
	if _, werr := w.Write(img); werr != nil {
		return nil, werr
	}

	// Close encoder for writing
	// log.Printf("Closing writer\n")
	if cerr := w.Close(); cerr != nil {
		return nil, cerr
	}

	// Return final encoded bytes
	// log.Printf("Length: %v\n", byt.Len())
	return byt.Bytes(), nil
}

// Decompress images with zlib
func DecodeImage(encimg []byte) ([]byte, error) {

	// Create buffer reader from bytes
	byt := bytes.NewReader(encimg)

	// Create zlib processor
	r, err := zlib.NewReader(byt)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Decode all bytes into new []byte
	decbyt, derr := ioutil.ReadAll(r)
	if derr != nil {
		return nil, derr
	}

	return decbyt, nil
}
