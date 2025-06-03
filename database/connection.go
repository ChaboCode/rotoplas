package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DBInstance() *mongo.Client {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongodb := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(mongodb))

	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	log.Println("Connected to MongoDB")
	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(client *mongo.Client, name string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("rotoplas").Collection(name)
	return collection
}
