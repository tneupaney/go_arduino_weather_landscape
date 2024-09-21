package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
)

type WeatherLandscape struct {
	OWM_KEY          string
	OWM_LAT, OWM_LON float64
	TMP_DIR          string
	OUT_FILENAME     string
	OUT_FILEEXT      string
	TEMPLATE_FILENAME string
	SPRITES_DIR      string
	DRAWOFFSET       int
}

func NewWeatherLandscape() *WeatherLandscape {
	wl := &WeatherLandscape{
		OWM_KEY:           "",  // Replace with your API key
		OWM_LAT:           52.196136,
		OWM_LON:           21.007963,
		TMP_DIR:           "tmp",
		OUT_FILENAME:      "test_",
		OUT_FILEEXT:       ".bmp",
		TEMPLATE_FILENAME: "p_weather/template.bmp",
		SPRITES_DIR:       "p_weather/sprite",
		DRAWOFFSET:        65,
	}

	if wl.OWM_KEY == "" {
		panic("Set OWM_KEY variable to your OpenWeather API key")
	}
	return wl
}

func (wl *WeatherLandscape) MakeImage() image.Image {
	// Replace this with actual OpenWeatherMap API call and processing logic
	fmt.Println("Fetching weather data from OpenWeatherMap")

	imgFile, err := os.Open(wl.TEMPLATE_FILENAME)
	if err != nil {
		panic("Failed to open template image")
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic("Failed to decode image")
	}

	// Simulate drawing weather data on the image
	fmt.Println("Drawing weather data on image")

	return img
}

func (wl *WeatherLandscape) SaveImage() string {
	img := wl.MakeImage()
	placekey := fmt.Sprintf("%.4f_%.4f", wl.OWM_LAT, wl.OWM_LON)
	outfilepath := wl.TmpFilePath(wl.OUT_FILENAME + placekey + wl.OUT_FILEEXT)

	file, err := os.Create(outfilepath)
	if err != nil {
		panic("Failed to create output file")
	}
	defer file.Close()

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		panic("Failed to save image")
	}
	return outfilepath
}

func (wl *WeatherLandscape) TmpFilePath(filename string) string {
	return filepath.Join(wl.TMP_DIR, filename)
}

func main() {
	wl := NewWeatherLandscape()
	fn := wl.SaveImage()
	fmt.Println("Saved", fn)
}
