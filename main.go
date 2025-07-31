package main

import (
	"fmt"
	"hash/fnv"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"text/template"

	"github.com/nfnt/resize"

	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"rotoplas/database"
	"rotoplas/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg":   true,
	"image/png":    true,
	"image/gif":    true,
	"image/webp":   true,
	"image/x-icon": true,
}

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, []string{"Album1", "Album2", "Album3"})
}

func getMimeFromHeader(header *multipart.FileHeader) string {
	return header.Header.Get("Content-Type")
}

func postFile(c *gin.Context) {
	hiddenTxt := c.Request.PostFormValue("hidden")
	log.Print(hiddenTxt)
	hidden, _ := strconv.ParseBool(hiddenTxt)
	log.Print(hidden)
	file, header, _ := c.Request.FormFile("file")
	log.Printf("%s", header.Filename)
	encoded := hash(header.Filename)
	encoded += "." + strings.Split(header.Filename, ".")[len(strings.Split(header.Filename, "."))-1]

	// Step 1. Save full file
	err := c.SaveUploadedFile(header, "./files/"+encoded)
	if err != nil {
		return
	}

	// Check for image

	if allowedImageTypes[header.Header.Get("Content-Type")] {
		err = processImage(file, header, encoded)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	res, err := database.AddFile(models.File{
		Name:      encoded,
		Size:      header.Size,
		CreatedAt: time.Now(),
		UploadIP:  c.ClientIP(),
		MimeType:  getMimeFromHeader(header),
		Hidden:    hidden,
	})

	if err != nil {
		c.String(http.StatusInternalServerError, "Error saving file metadata")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": encoded,
		"id":       res,
	})
}

func processImage(file io.Reader, header *multipart.FileHeader, name string) error {
	// Check for image MIME type
	if !allowedImageTypes[header.Header.Get("Content-Type")] {
		return fmt.Errorf("not img: %s", header.Header.Get("Content-Type"))
	}

	if _, err := os.Stat("files/thumbs"); os.IsNotExist(err) {
		if err := os.Mkdir("files/thumbs", os.ModePerm); err != nil {
			return fmt.Errorf("failed to create thumbs dir: %v", err)
		}
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	ratio := float32(img.Bounds().Max.X) / float32(img.Bounds().Max.Y)

	thumb := resize.Resize(uint(ratio*150), 150, img, resize.Bilinear)
	thumbFile, err := os.Create("files/thumbs/" + name)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail file: %v", err)
	}
	defer thumbFile.Close()

	if err := jpeg.Encode(thumbFile, thumb, nil); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %v", err)
	}

	return nil
}

func faqPage(c *gin.Context) {
	c.HTML(http.StatusOK, "faq.html", gin.H{})
}

type File struct {
	Name  string
	IsImg bool
	IsVid bool
	Size  string
}

func getHash(c *gin.Context) {
	data := c.Query("data")
	if data == "" {
		c.String(http.StatusBadRequest, "Missing 'data' query parameter")
		return
	}

	c.String(http.StatusOK, hash(data))
}

func hash(data string) string {
	h := fnv.New32a()
	h.Write([]byte(data))
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}

func getThumbnail(c *gin.Context) {
	filename := c.Param("filename")
	c.Header("Content-Type", "image/jpeg")
	c.File("./files/thumbs/" + filename)
}

func isImg(mime string) bool {
	return strings.HasPrefix(mime, "image/")
}

func isVid(mime string) bool {
	return strings.HasPrefix(mime, "video/")
}

func isAudio(mime string) bool {
	return strings.HasPrefix(mime, "audio/")
}

func homePage(c *gin.Context) {
	pageStr := c.Query("page")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1 // Default to page 1 if parsing fails
	}

	files, err := database.ListFiles(10, page)
	count, _ := database.Count()

	if err != nil {
		log.Printf("Error listing files: %v", err)
		c.String(http.StatusInternalServerError, "Error retrieving files")
		return
	}
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"files":    files,
		"admin":    false,
		"next":     int(count) > page*10,
		"prev":     page > 1,
		"nextPage": page + 1,
		"prevPage": page - 1})
}

func goGodMode(c *gin.Context) {
	pageStr := c.Query("page")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1 // Default to page 1 if parsing fails
	}

	files, err := database.ListGodMode(10, page)
	count, _ := database.Count()

	if err != nil {
		log.Printf("Error listing files: %v", err)
		c.String(http.StatusInternalServerError, "Error retrieving files")
		return
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"files":    files,
		"admin":    true,
		"next":     int(count) > page*10,
		"prev":     page > 1,
		"nextPage": page + 1,
		"prevPage": page - 1})

}

func deleteFile(c *gin.Context) {
	name := c.Query("name")

	err := database.DeleteFile(name)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting file")
		return
	}

	c.Redirect(http.StatusSeeOther, "/godmode")
}

func main() {
	godotenv.Load()

	database.ConnectMySQL()

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"isImg":   isImg,
		"isVid":   isVid,
		"isAudio": isAudio,
	})
	router.LoadHTMLGlob("templates/*")
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	// router.GET("/", homePage)
	router.GET("/faq", faqPage)
	router.GET("/api/files/list", func(c *gin.Context) {
		countStr := c.Query("count")
		pageStr := c.Query("page")

		count, err1 := strconv.Atoi(countStr)
		page, err2 := strconv.Atoi(pageStr)

		if err1 != nil || err2 != nil {
			c.String(http.StatusBadRequest, "Invalid count or page parameter")
			return
		}

		data, _ := database.ListFiles(count, page)
		c.JSON(http.StatusOK, gin.H{
			"files": data,
		})
	})

	router.GET("/", homePage)
	router.GET("/godmode", goGodMode)
	router.GET("/hash", getHash)
	router.POST("/upload", postFile)
	router.GET("/delete/", deleteFile)
	router.Static("/files", "./files")
	router.GET("/thumbs/:filename", getThumbnail)
	router.StaticFile("favicon.ico", "./favicon.ico")
	router.StaticFile("bg.png", "./bg.png")
	router.StaticFile("rotoplas.png", "./rotoplas.png")
	// router.Use(middleware.Auth())

	router.Run("0.0.0.0:" + os.Getenv("PORT"))
}
