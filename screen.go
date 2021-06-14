package main

import (
	"bytes"
	"compress/zlib"
	"time"

	"github.com/cretz/go-scrap"
)

// Get a list of capturers for all displays available
func GetAllCapturers() ([]*scrap.Capturer, error) {

	// Get all displays
	disps, err := scrap.Displays()
	if err != nil {
		panic(err)
	}

	// store capturers
	caps := make([]*scrap.Capturer, len(disps))

	// Loop over displays
	for i, disp := range disps {

		// Create capturer for it
		cap, err := scrap.NewCapturer(disp)
		if err != nil {
			return nil, err
		}

		caps[i] = cap
	}

	return caps, nil
}

// Capture a screenshot for a screen capturer
func CaptureScreen(c *scrap.Capturer) (*scrap.FrameImage, error) {

	// Get an image, trying until one available
	for {
		if img, _, err := c.FrameImage(); img != nil || err != nil {
			// Detach the image so it's safe to use after this method
			if img != nil {
				img.Detach()
			}
			return img, err
		}

		// Sleep 17ms (~1/60th of a second) and try again
		time.Sleep(17 * time.Millisecond)
	}
}

// Compress screenshot with zlib
func EncodeImage(img *scrap.FrameImage) []byte {

	// Create buffer and encode image
	var byt bytes.Buffer
	w := zlib.NewWriter(&byt)
	w.Write(img.Pix)

	return byt.Bytes()
}
