package helper

import (
	"context"
	"log"
	"os"
	"rotoplas/database"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SignedDetails struct {
	Email string
	Uid   string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, userId string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email: email,
		Uid:   userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 15000, // Set expiration time as needed
			Issuer:    "rotoplas",
		},
	}

	refreshClaims := &SignedDetails{
		Email: email,
		Uid:   userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 15000, // Set expiration time as needed
			Issuer:    "rotoplas",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err = token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err = refreshToken.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return signedToken, signedRefreshToken, nil
}

func ValidateToken(signedToken string) (claims *SignedDetails, status bool) {
	token, _ := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, false
	}

	return claims, true
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var updateObj primitive.D
	updateObj = append(updateObj, primitive.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, primitive.E{Key: "refresh_token", Value: signedRefreshToken})

	filter := bson.M{"user_id": userId}

	_, err := userCollection.UpdateOne(context.TODO(), filter, bson.D{
		{Key: "$set", Value: updateObj}}, options.UpdateOne())

	if err != nil {
		log.Panic(err)
		return
	}

	return
}
