package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

const (
	DEFAULTSIZE int = 100
	ASCII_CHARS  = "@%#*+=-:. "
	MAX_IMAGE_SIZE int64  = 8 << 20
)

type ConvertParams struct {
    Size    int    `json:"size" binding:"gte=1,lte=300"`
    CharSet string `json:"charSet" binding:"required,min=2,max=32"`
}

var logFile *os.File

func init(){
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err!=nil{fmt.Println("Error opening log file:", err); os.Exit(1)}
	
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, logFile)
}

func convertHandler(c *gin.Context){
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error":"File upload required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	if _, err = file.Read(buf); err != nil {
		c.JSON(400, gin.H{"error": "Failed to read file header"})
		return
	}

	if _, err = file.Seek(0, 0); err != nil {
		c.JSON(500, gin.H{"error": "Failed to reset file pointer"})
		return
	}

	mimeType := http.DetectContentType(buf)
	if !allowedTypes[mimeType] {
		c.JSON(400, gin.H{"error": "Unsupported image type"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
    defer cancel()

	jsonData := c.PostForm("params")
    var params ConvertParams
    if err := json.Unmarshal([]byte(jsonData), &params); err != nil {
        c.JSON(400, gin.H{"error": "Invalid JSON format: " + err.Error()})
        return
    }

	ascii, err := convertToASCII(ctx, file, fileHeader.Filename, params)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
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
	if x < y {
		scale = float64(y) / float64(size)
	}
	return resize.Resize(uint(float64(x)/scale), uint(float64(y)/scale),img, resize.Lanczos3)
}

func convertToGray(img image.Image) *image.Gray {
    bounds := img.Bounds()
    gray := image.NewGray(bounds)
    draw.Draw(gray, bounds, img, bounds.Min, draw.Src)
    return gray
}

func convertToASCII(ctx context.Context, file io.Reader, filename string, params ConvertParams) (string, error){
	ext := strings.ToLower(filepath.Ext(filename))
    var decodeFunc func(io.Reader) (image.Image, error)

    switch ext {
    case ".jpg", ".jpeg":
        decodeFunc = func(r io.Reader) (image.Image, error) {
            return jpeg.Decode(r)
        }
    case ".png":
        decodeFunc = func(r io.Reader) (image.Image, error) {
            return png.Decode(r)
        }
    default:
        return "", fmt.Errorf("unsupported file format: %s", ext)
    }

    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }

    img, err := decodeFunc(io.LimitReader(file, MAX_IMAGE_SIZE))
    if err != nil {
        return "", fmt.Errorf("failed to decode image: %v", err)
    }

	bounds := img.Bounds()
	if bounds.Dx() > 5000 || bounds.Dy() > 5000 {
		return "", fmt.Errorf("image dimensions too large")
	}

	if params.Size > 300 || params.Size <= 0 {params.Size = DEFAULTSIZE}
	if utf8.RuneCountInString(params.CharSet) < 2 || utf8.RuneCountInString(params.CharSet) > 32 {params.CharSet = ASCII_CHARS}

	img = resizeImage(img, params.Size)
	gray := convertToGray(img)

	select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }

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
	gin.SetMode(gin.ReleaseMode)
	log.SetOutput(io.Discard)
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

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    <-quit
    fmt.Println("Shutting down server...")

    if logFile != nil {logFile.Close()}

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    fmt.Println("Server exiting")
}