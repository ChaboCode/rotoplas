package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"rotoplas/controllers"
	"rotoplas/database"
	"rotoplas/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, []string{"Album1", "Album2", "Album3"})
}

func postFile(c *gin.Context) {
	fileCollection := database.OpenCollection(database.Client, "files")

	file, _ := c.FormFile("file")
	log.Printf("%s", file.Filename)
	encoded := hash(file.Filename)
	encoded += "." + strings.Split(file.Filename, ".")[len(strings.Split(file.Filename, "."))-1]

	res, err := fileCollection.InsertOne(c, models.File{
		Name:      encoded,
		Size:      file.Size,
		CreatedAt: time.Now(),
		UploadIP:  c.ClientIP(),
	})

	if err != nil {
		c.String(http.StatusInternalServerError, "Error saving file metadata")
		return
	}

	err = c.SaveUploadedFile(file, "./files/"+encoded)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": encoded,
		"id":       res.InsertedID,
	})
}

func homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"files": listFilesMongo(),
	})
}

type File struct {
	URL   string
	Name  string
	IsImg bool
	IsVid bool
	Size  string
}

func listFilesMongo() []File {
	fileCollection := database.OpenCollection(database.Client, "files")

	cursor, err := fileCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var files []File
	for cursor.Next(context.TODO()) {
		var file models.File
		if err := cursor.Decode(&file); err != nil {
			log.Fatal(err)
		}

		files = append(files, File{
			Name:  file.Name,
			URL:   "files/" + file.Name,
			IsVid: strings.TrimPrefix(path.Ext(file.Name), ".") == "mp4" || strings.TrimPrefix(path.Ext(file.Name), ".") == "webm",
			IsImg: strings.TrimPrefix(path.Ext(file.Name), ".") == "jpg" || strings.TrimPrefix(path.Ext(file.Name), ".") == "jpeg" || strings.TrimPrefix(path.Ext(file.Name), ".") == "png" || strings.TrimPrefix(path.Ext(file.Name), ".") == "gif" || strings.TrimPrefix(path.Ext(file.Name), ".") == "webp" || strings.TrimPrefix(path.Ext(file.Name), ".") == "avif" || strings.TrimPrefix(path.Ext(file.Name), ".") == "svg",
			Size:  fmt.Sprintf("%.5f", float32(file.Size)/(1024*1024)),
		})
	}

	return files
}

func listFiles(dir string) []File {
	root := os.DirFS(dir)

	mdFiles, err := fs.Glob(root, "*")

	if err != nil {
		log.Fatal(err)
	}

	var files []File
	for _, v := range mdFiles {
		files = append(files, File{
			Name:  path.Join(dir, v),
			IsVid: strings.TrimPrefix(path.Ext(v), ".") == "mp4" || strings.TrimPrefix(path.Ext(v), ".") == "webm",
			IsImg: strings.TrimPrefix(path.Ext(v), ".") == "jpg" || strings.TrimPrefix(path.Ext(v), ".") == "jpeg" || strings.TrimPrefix(path.Ext(v), ".") == "png" || strings.TrimPrefix(path.Ext(v), ".") == "gif" || strings.TrimPrefix(path.Ext(v), ".") == "webp" || strings.TrimPrefix(path.Ext(v), ".") == "avif" || strings.TrimPrefix(path.Ext(v), ".") == "svg",
		})
	}
	return files
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

func main() {
	godotenv.Load()

	//mongoURI := os.Getenv("MONGO_URI")
	// client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	// if err != nil {
	// panic(err)
	// }

	// defer func() {
	// 	if err := client.Disconnect(nil); err != nil {
	// 		panic(err)
	// 	}
	// }()

	// coll := client.Database("rotoplas").Collection("users")

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.POST("/signup", controllers.SignUp())
	router.POST("/login", controllers.Login())
	router.GET("/", homePage)

	router.GET("/hash", getHash)
	router.POST("/upload", postFile)
	router.Static("/files", "./files")
	router.StaticFile("favicon.ico", "./favicon.ico")
	// router.Use(middleware.Auth())

	router.Run(":" + os.Getenv("PORT"))
}
