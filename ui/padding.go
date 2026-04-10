package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"threatintel-feed-wizard/images"
)

// ScreenMargin controls the padding around every wizard screen.
const ScreenMargin float32 = 20

// padded wraps content with consistent margins on all sides.
func padded(content fyne.CanvasObject) fyne.CanvasObject {
	top := canvas.NewRectangle(nil)
	top.SetMinSize(fyne.NewSize(0, ScreenMargin))
	bottom := canvas.NewRectangle(nil)
	bottom.SetMinSize(fyne.NewSize(0, ScreenMargin))
	left := canvas.NewRectangle(nil)
	left.SetMinSize(fyne.NewSize(ScreenMargin, 0))
	right := canvas.NewRectangle(nil)
	right.SetMinSize(fyne.NewSize(ScreenMargin, 0))
	return container.NewBorder(top, bottom, left, right, content)
}

// screenImage loads an embedded image and returns it sized for the
// upper-right corner of a wizard screen.
func screenImage(filename string) fyne.CanvasObject {
	data, err := images.FS.ReadFile(filename)
	if err != nil {
		log.Printf("[ui] could not load embedded image %s: %v", filename, err)
		return container.NewWithoutLayout()
	}
	res := fyne.NewStaticResource(filename, data)
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(160, 160))
	return img
}
