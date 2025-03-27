package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

var MAXSIZE int = 100
const ASCII_CHARS  = "@%#*+=-:. "


func convertHandler(c *gin.Context){
	data, err := c.GetRawData()
	if err != nil{
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ascii, err := convertToASCII(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"ascii": ascii})
}

func resizeImage(img image.Image) image.Image{
	x, y := img.Bounds().Dx(), img.Bounds().Dx()
	scale := float64(x)/float64(MAXSIZE)
	if x<y {
		scale = float64(y)/float64(MAXSIZE)
	}
	return resize.Resize(uint(float64(x)/scale), uint(float64(y)/scale),img, resize.Lanczos3)
}

func convertToGray(img image.Image) *image.Gray {
    bounds := img.Bounds()
    gray := image.NewGray(bounds)
    draw.Draw(gray, bounds, img, bounds.Min, draw.Src)
    return gray
}

func convertToASCII(data []byte) (string, error){
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err!=nil{return "", err}
	img = resizeImage(img)
	gray := convertToGray(img)
	return generateASCII(gray), nil

}

func generateASCII(img *image.Gray) string {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	var asciiArt bytes.Buffer

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			char := brightnessToASCII(float64(img.GrayAt(x, y).Y) / 255.0)
			asciiArt.WriteByte(char)
		}
		asciiArt.WriteByte('\n')
	}

	return asciiArt.String()
}

func brightnessToASCII(b float64) byte {
	index := int(b * float64(len(ASCII_CHARS)-1))
	return ASCII_CHARS[index]
}

func aboutHandler(c *gin.Context){
	c.JSON(200, gin.H{
		"about": "Project of image translation to ASCI code",
		"tip" :"For the algorithm to work more correctly, contrasting images are required.",
	})
}

func main(){
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	router.POST("/convert", convertHandler)
	router.GET("/about", aboutHandler)
	
    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout: 60 * time.Second,
    }

    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}