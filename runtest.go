package main

import (
	"fmt"
)

type WeatherLandscape struct{}

func (wl *WeatherLandscape) SaveImage() string {
	// Placeholder logic for saving an image
	fileName := "test_image.jpg"
	fmt.Println("Image saved:", fileName)
	return fileName
}

func main() {
	w := WeatherLandscape{}
	fileName := w.SaveImage()
	fmt.Println("Saved", fileName)
}
