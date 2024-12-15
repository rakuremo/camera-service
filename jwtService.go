package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		log.Fatal("Error loading .env file")
	}
}

func generateJWT(username string) (string, error) {
	var jwtSecret = []byte("SecretYouShouldHide")
	if string(jwtSecret) == "" {
		return "", fmt.Errorf("JWT secret not set in .env")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour)
	claims["username"] = username

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Debug(err)
		return "", err
	}

	return tokenString, nil
}
