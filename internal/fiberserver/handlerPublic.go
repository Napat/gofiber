package fiberserver

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func helloWorldHandler(c *fiber.Ctx) error {
	myPrinter(os.Stdout, "Hello, World!\n")
	return c.SendString("Hello, World!")
}

func myPrinter(writer io.Writer, s string) {
	fmt.Fprintf(writer, "MyPrinter: %s", s)
}

// loginHandler fiber httpHandler to receive user/password via FormValue data
// Response JWT token when successful
func loginHandler(c *fiber.Ctx) error {
	user := c.FormValue("user")
	pass := c.FormValue("pass")

	// Validate credential and throws error if unauthorized
	if user != "john" || pass != "doe" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create jwt token
	token := jwt.New(jwt.SigningMethodRS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "John Doe"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	privateKey, err := jwtCredGet(mykey)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Not found key provided",
		})
	}

	// Generate encoded token and send it as response.
	t, err := token.SignedString(privateKey)
	if err != nil {
		log.Printf("token.SignedString: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}
