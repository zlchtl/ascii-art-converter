package main

import (
	"bytes"
	"context"
	"image"
	"image/draw"
	"io"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

const (
	DEFAULTSIZE int = 100
	ASCII_CHARS  = "@%#*+=-:. "
)

type ConvertParams struct {
    Size int `json:"Size" binding:"gte=1,lte=300"`
	CharSet string `json:"charSet" binding:"required,min=2,max=32"`
}

func convertHandler(c *gin.Context){
	fileHeader, err := c.FormFile("file")
	if err != nil {c.JSON(400, gin.H{"error":"File upload required"}); return}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("Error processing image: %v", err)
		c.JSON(500, gin.H{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
    defer cancel()

	var params ConvertParams
	if err := c.ShouldBind(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ascii, err := convertToASCII(ctx, file, params)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Timeout processing image after 20s")
            c.JSON(504, gin.H{"error": "Processing timeout"})
        } else {
            c.JSON(500, gin.H{"error": err.Error()})
        }
        return
	}

	c.JSON(200, gin.H{"ascii": ascii})
	c.Header("Content-Type", "application/json; charset=utf-8")
}

func resizeImage(img image.Image, size int) image.Image{
	x, y := img.Bounds().Dx(), img.Bounds().Dy()
	scale := float64(x)/float64(size)
	if x<y {
		scale = float64(y)/float64(size)
	}
	return resize.Resize(uint(float64(x)/scale), uint(float64(y)/scale),img, resize.Lanczos3)
}

func convertToGray(img image.Image) *image.Gray {
    bounds := img.Bounds()
    gray := image.NewGray(bounds)
    draw.Draw(gray, bounds, img, bounds.Min, draw.Src)
    return gray
}

func convertToASCII(ctx context.Context, file io.Reader, params ConvertParams) (string, error){
	select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }
	
	img, _, err := image.Decode(io.LimitReader(file, 10<<20))
	if err!=nil{return "", err}

	select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }

	if params.Size > 300 || params.Size <= 0 {params.Size = DEFAULTSIZE}
	if utf8.RuneCountInString(params.CharSet) < 2 || utf8.RuneCountInString(params.CharSet) > 32 {params.CharSet = ASCII_CHARS}

	img = resizeImage(img, params.Size)
	gray := convertToGray(img)
	return generateASCII(gray, params.CharSet), nil
}

func generateASCII(img *image.Gray, charset string) string {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	var asciiArt bytes.Buffer

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			charsetLen := float64(len(charset)-1)
			brightness := float64(img.GrayAt(x, y).Y)/255.0
			char := charset[int(brightness * charsetLen)]
			asciiArt.WriteByte(char)
		}
		asciiArt.WriteByte('\n')
	}

	return asciiArt.String()
}

func aboutHandler(c *gin.Context){
	c.JSON(200, gin.H{
		"about": "Project of image translation to ASCI code",
		"tip" :"For the algorithm to work more correctly, contrasting images are required.",
	})
}

func main(){
	//gin.SetMode(gin.ReleaseMode)
	//log.SetOutput(io.Discard)
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