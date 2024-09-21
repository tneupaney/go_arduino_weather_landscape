package main

import (
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"image/jpeg"
	"weatherlandscape" // Import your package that contains WeatherLandscape
)

const (
	SERV_IPADDR      = "0.0.0.0"
	SERV_PORT        = 3355
	EINKFILENAME     = "test.bmp"
	USERFILENAME     = "test1.bmp"
	FILETOOOLD_SEC   = 60 * 10
)

var WEATHER = weatherlandscape.NewWeatherLandscape() // Assuming a constructor

func isFileTooOld(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return true
	}
	return time.Since(info.ModTime()) > FILETOOOLD_SEC*time.Second
}

func createWeatherImages() {
	userFileName := WEATHER.TmpFilePath(USERFILENAME)
	einkFileName := WEATHER.TmpFilePath(EINKFILENAME)

	if !isFileTooOld(userFileName) {
		return
	}

	img := WEATHER.MakeImage() // Assuming MakeImage returns an image.Image

	// Save image to file (JPEG as an example)
	file, _ := os.Create(userFileName)
	defer file.Close()
	jpeg.Encode(file, img, nil)

	// Rotate and flip the image if required
	// Save the transformed image as einkFileName
}

func indexHtml() string {
	body := "<h1>Weather as Landscape</h1>"
	body += fmt.Sprintf("<p>Place: %.4f, %.4f</p>", WEATHER.Lat, WEATHER.Lon)
	body += "<p><img src=\"" + USERFILENAME + "\" alt=\"Weather\"></p>"
	body += "<p>ESP32 URL: <span id=\"eink\"></span></p>"
	body += "<script>document.getElementById(\"eink\").innerHTML = window.location+\"" + EINKFILENAME + "\";</script>"

	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		  <head>
		    <meta charset="utf-8">
		    <title>Weather as Landscape</title>
		  </head>
		  <body>%s</body>
		</html>`, body)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}

	if r.URL.Path == "/index.html" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, indexHtml())
		return
	}

	if r.URL.Path == "/"+EINKFILENAME || r.URL.Path == "/"+USERFILENAME {
		createWeatherImages()

		fileName := WEATHER.TmpFilePath(r.URL.Path[1:])
		file, err := os.Open(fileName)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeFile(w, r, fileName)
	}
}

func main() {
	http.HandleFunc("/", handler)
	address := fmt.Sprintf("%s:%d", SERV_IPADDR, SERV_PORT)
	fmt.Printf("Serving at http://%s/\n", address)
	http.ListenAndServe(address, nil)
}
