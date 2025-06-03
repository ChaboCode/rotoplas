package main

import (
	"hash/fnv"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"rotoplas/controllers"
	"rotoplas/middleware"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, []string{"Album1", "Album2", "Album3"})
}

func postFile(c *gin.Context) {
	file, _ := c.FormFile("file")
	log.Printf("%s", file.Filename)
	encoded := hash(file.Filename)
	encoded += "." + strings.Split(file.Filename, ".")[1]

	c.SaveUploadedFile(file, "./files/"+encoded)
	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": encoded,
	})
}

func homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"imgs": listFiles("./files"),
	})
}

func listFiles(dir string) []string {
	root := os.DirFS(dir)

	mdFiles, err := fs.Glob(root, "*")

	if err != nil {
		log.Fatal(err)
	}

	var files []string
	for _, v := range mdFiles {
		files = append(files, path.Join(dir, v))
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
	router.Use(middleware.Auth())

	router.GET("/hash", getHash)
	router.POST("/upload", postFile)
	router.Static("/files", "./files")

	router.Run(":" + os.Getenv("PORT"))
}
