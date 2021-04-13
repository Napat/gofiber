package main

import (
	"crypto/rand"
	"crypto/rsa"

	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django"

	jwt "github.com/form3tech-oss/jwt-go"
	jwtware "github.com/gofiber/jwt/v2"

	"github.com/go-playground/validator/v10"
)

type TransID int

type ReqHeader struct {
	TID TransID `json:"id"`
}

type Greet struct {
	//ReqHeader

	Uid   TransID `json:"uid" validate:"required"`
	Title string  `json:"title" validate:"required"`
	Msg   string  `json:"message" validate:"required"`
}

type GreetV2 struct {
	UnixNano int64   `json:"utime" validate:"required"`
	Greets   []Greet `json:"greets" validate:"dive"`
}

var greets = map[TransID]Greet{
	10000001: {
		Uid:   1,
		Title: "Hello Foo",
		Msg:   "Hello To Mr.Foo",
	},
	10000002: {
		Uid:   2,
		Title: "Hello Bar",
		Msg:   "Hello To Ms.Bar",
	},
}

type ReqUserInfo struct {
	TID      TransID `json:"tid" validate:"required"`
	Name     string  `json:"name" validate:"required,min=3,max=32"`
	IsActive bool    `json:"isactive" validate:"required,eq=True|eq=False"`
	Email    string  `json:"email" validate:"required,email,min=6,max=32"`
	Job      Job     `json:"job" validate:"dive"`
}

type Job struct {
	Type   string `json:"type" validate:"required,min=3,max=32"`
	Salary int    `json:"salary" validate:"required,number"`
}

type RespUserInfo struct {
	TID       TransID     `json:"tid"`
	Name      string      `json:"name"`
	UserToken interface{} `json:"isactive"`
}

type RespError struct {
	FailedField string
	Tag         string
	Value       string
}

// jwt private key
var privateKey *rsa.PrivateKey

// Nl2BrHtml Custom function for template
func Nl2BrHtml(value interface{}) string {
	if str, ok := value.(string); ok {
		return strings.Replace(str, "\n", "<br />", -1)
	}
	return ""
}

func NewFiber(enableTemplate bool) *fiber.App {

	if !enableTemplate {
		app := fiber.New()
		return app
	} else {
		/////////////////////// template[django]
		// https://ichi.pro/th/reiyn-ru-keiyw-kab-fiber-krxb-kar-phat-hna-web-golang-him-103546697181187
		// https://github.com/gofiber/template
		// django: https://github.com/gofiber/template/tree/master/django
		engine := django.New("./web/django/views", ".django")
		// register functions
		engine.AddFunc("nl2br", Nl2BrHtml)
		app := fiber.New(fiber.Config{Views: engine})
		return app
	}
}

func jwtPrivateKey() (*rsa.PrivateKey, error) {
	// Just as a demo, generate a new private/public key pair on each run. DO NOT DO THIS IN PRODUCTION!
	rng := rand.Reader
	var err error
	privateKey, err = rsa.GenerateKey(rng, 2048)
	if err != nil {
		log.Fatalf("rsa.GenerateKey: %v", err)
		return nil, err
	}

	return privateKey, nil
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

	// Generate encoded token and send it as response.
	t, err := token.SignedString(privateKey)
	if err != nil {
		log.Printf("token.SignedString: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

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

func ValidateStruct(userInfo interface{}) []RespError {
	var errors []RespError
	validate := validator.New()
	err := validate.Struct(userInfo)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element RespError
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, element)
		}
	}
	return errors
}

func main() {
	//// Step 1: Create fiber
	app := NewFiber(true)

	//// Step 2: Create Grouping: https://docs.gofiber.io/guide/routing#grouping
	api := app.Group("/api")         // /api
	apiv1 := api.Group("/v1")        // /api/v1
	apiv2 := api.Group("/v2")        // /api/v2
	private := app.Group("/private") // /private

	//// Step 3: Init Middleware

	// Internal middleware
	// middleware logger: https://github.com/gofiber/fiber/blob/master/middleware/logger/README.md
	app.Use(logger.New(logger.Config{
		Format:     "${blue}${time} ${pid} ${red}${status} ${method} ${yellow}${path} ${green}${ip} ${ua} ${bytesSent} ${white}${resBody}${reset}\n",
		TimeFormat: "2006/Jan/02",
		TimeZone:   "Asia/Bangkok",
	}))

	// Internal middleware: cache https://github.com/gofiber/fiber#-internal-middleware
	// curl -v http://127.0.0.1:3000/api/v2/greets
	// curl -v http://127.0.0.1:3000/api/v2/greets?refresh=true
	apiv2.Use(cache.New(cache.Config{
		Next:         nil,
		Expiration:   30 * time.Second,
		CacheControl: true,
		KeyGenerator: func(c *fiber.Ctx) string {
			if c.Query("refresh") == "true" {
				return c.Path() + time.Now().String()
			}
			return c.Path()
		},
		Storage: nil,
	}))

	// ExternalMiddleware: https://github.com/gofiber/fiber#-external-middleware
	// JWT https://github.com/gofiber/jwt
	// JWT Middleware
	// **NOTE** Must call .Use(...jwtware...) before register route/method/handler, i.e. private.Get("/", privateHandler)
	privateKey, err := jwtPrivateKey()
	if err != nil {
		return
	}

	private.Use(jwtware.New(jwtware.Config{
		SigningMethod: "RS256",
		SigningKey:    privateKey.Public(),
	}))

	//// Step 4: Init Enpoint / Method

	// GET plaintext
	// curl -v http://127.0.0.1:3000/
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// GET Response plaintext with parameters: https://docs.gofiber.io/guide/routing#parameters
	// curl -v http://127.0.0.1:3000/user/foo/books/the_intelligent_investor?page=111
	app.Get("/user/:name/books/:title", func(c *fiber.Ctx) error {
		c.Set("Connection", "keep-alive")
		c.Set("Content-Type", "text/html; charset=utf-8")

		return c.SendString("name: " + c.Params("name") + " title: " + c.Params("title") + " page: " + c.Query("page"))
	})

	// GET Response JSON
	// curl -sv http://127.0.0.1:3000/health | jq
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(&fiber.Map{
			"status": "ok",
		})
	})

	// GET Response JSON
	// curl -sv http://127.0.0.1:3000/api/v1/greets | jq
	apiv1.Get("/greets", func(c *fiber.Ctx) error {
		var result []Greet
		for _, greet := range greets {
			result = append(result, greet)
		}

		log.Printf("Response json: %v\n", result)
		return c.Status(http.StatusOK).JSON(result)
	})

	// curl -sv http://127.0.0.1:3000/api/v2/greets | jq
	apiv2.Get("/greets", func(c *fiber.Ctx) error {
		var result GreetV2
		for _, greet := range greets {
			result.Greets = append(result.Greets, greet)
		}

		result.UnixNano = time.Now().UnixNano()

		log.Printf("Response json: %v\n", result)
		return c.Status(http.StatusOK).JSON(result)
	})

	// Static files Server
	// curl http://127.0.0.1:3000/web
	app.Static("/web", "./web/static")

	// template[django]
	// curl http://127.0.0.1:3000/template/django/
	app.Get("/template/django/", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	// template[django]
	// curl http://127.0.0.1:3000/template/django/layout
	app.Get("/template/django/layout", func(c *fiber.Ctx) error {
		// Render index within layouts/main
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layouts/main")
	})

	// Example using Middleware JWT: Recv Form data(user/password)
	// # Login / Get token
	// - curl --data "user=john&pass=doe" http://localhost:3000/login
	// - TOKEN=$(curl -s --data "user=john&pass=doe" http://localhost:3000/login | grep token | cut -d\" -f4); echo $TOKEN
	// # Access private route
	// - curl -v localhost:3000/private -H "Authorization: Bearer $TOKEN"
	app.Post("/login", loginHandler)
	private.Get("/", privateHandler)

	// Using Validator to validate the request struct
	// -[Fail case] curl -vs -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" --data "{\"name\":\"john\",\"isactive\":true}" http://localhost:3000/private/userinfo | jq
	// -[Success case] curl -vs -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" --data '{"tid":123456789,"name":"john","isactive":true,"email":"xxx@xxx.com","job":{"type":"salaryman","salary":10000}}' http://localhost:3000/private/userinfo | jq
	private.Post("/userinfo", userInfoHandler)

	// Graceful shutdown
	// Listen from a different goroutine
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

	_ = <-c // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	fmt.Println("Running cleanup tasks...")

	// Your cleanup tasks go here
	// db.Close()
	// redisConn.Close()
	fmt.Println("Fiber was successful shutdown.")
}
