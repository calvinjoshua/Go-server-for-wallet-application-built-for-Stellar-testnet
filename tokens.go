package main

import (

	// "os"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func login(c *fiber.Ctx) error {
	payload := struct {
		Name string `json:"userName"`
		Mpin string `json:"mpin"`
	}{}

	if err := c.BodyParser(&payload); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	userName := payload.Name
	mpin := payload.Mpin

	if userName != "diam" || mpin != "diam123" {
		return c.JSON(fiber.Map{"Status": "401", "message": "Unauthorized"})
	}

	claims := jwt.MapClaims{
		"name": userName,
		"type": "Access token",
		"exp":  time.Now().Add(time.Minute * 5).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	accessToken, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	claims1 := jwt.MapClaims{
		"name": userName,
		"type": "Refresh token",
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}

	// Create token
	token1 := jwt.NewWithClaims(jwt.SigningMethodHS256, claims1)

	// Generate encoded token and send it as response.
	refreshToken, err := token1.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	fmt.Println(accessToken)

	return c.JSON(fiber.Map{"statusCode": 200, "accessToken": accessToken, "refreshToken": refreshToken})
}

func Regenerate(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)

	// Create the Claims
	claims1 := jwt.MapClaims{
		"name": name,
		"type": "Access token",
		"exp":  time.Now().Add(time.Minute * 1).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims1)

	// Generate encoded token and send it as response.
	accessToken, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"statusCode": 200, "accessToken": accessToken})
}
