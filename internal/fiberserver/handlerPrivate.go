package fiberserver

import (
	"fmt"

	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func privateHandler(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.SendString("Welcome " + name)
}

func userInfoHandler(c *fiber.Ctx) error {
	req := new(ReqUserInfo)
	if err := c.BodyParser(req); err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
		return err
	}

	errors := ValidateStruct(req)
	if errors != nil {
		fmt.Println(errors)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": false, "errorinfo": errors})
	}

	fmt.Println("0004")
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)

	// return c.JSON(fiber.Map{
	// 	"name":      name,
	// 	"usertoken": user,
	// })

	var resp RespUserInfo
	resp.TID = req.TID
	resp.Name = name
	resp.UserToken = user

	return c.JSON(resp)
}
