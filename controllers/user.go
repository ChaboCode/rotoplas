package controllers

import (
	"fmt"
	"log"
	"net/http"
	"rotoplas/database"
	"rotoplas/helper"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"rotoplas/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("user password is not valid")
		check = false
	}
	return check, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(c, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user count"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(c, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user count"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, user.UserID)
		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(c, user)
		if insertErr != nil {
			log.Panic(insertErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting user"})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		err := userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding user"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*foundUser.Password, *user.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, foundUser.UserID)
		helper.UpdateAllTokens(token, refreshToken, foundUser.UserID)
		c.JSON(http.StatusOK, foundUser)
	}
}
