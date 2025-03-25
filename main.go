package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func convert(c *gin.Context){
	data, err := c.GetRawData()
	if err != nil{
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = os.WriteFile("test.jpg", data, 0644)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Done": "Done"})
}

func about(c *gin.Context){
	c.JSON(200, gin.H{
		"about": "Project of image translation to ASCI code",
		"tip" :"For the algorithm to work more correctly, contrasting images are required.",
	})
}

func main(){
	router := gin.Default()
	router.POST("/convert", convert)
	router.GET("/about", about)
	router.Run(":8080")
}