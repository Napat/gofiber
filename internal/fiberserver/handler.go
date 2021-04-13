package fiberserver

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django"

	//jwt "github.com/form3tech-oss/jwt-go"
	jwtware "github.com/gofiber/jwt/v2"
)

type Handler struct {
	//// Step 1: Create fiber
	App *fiber.App

	//// Step 2: Create Grouping: https://docs.gofiber.io/guide/routing#grouping
	Api     fiber.Router
	Apiv1   fiber.Router
	Apiv2   fiber.Router
	Private fiber.Router
}

func NewServHandler() *Handler {
	h := Handler{}

	//// fiber Step 1: Create fiber
	h.initFiberApp(true)

	//// fiber Step 2: Create Grouping: https://docs.gofiber.io/guide/routing#grouping
	h.initFiberAppGroup()

	//// fiber Step 3: Init Middleware
	h.initMiddleware()

	//// fiber Step 4: Init Enpoint / Method
	h.initRoute()

	return &h
}

// Listen serves HTTP requests from the given addr.
//
//  App.Listen(":8080")
//  App.Listen("127.0.0.1:8080")
func (h *Handler) Listen(addr string) error {
	return h.App.Listen(addr)
}

func (h *Handler) initFiberApp(enableTemplate bool) error {
	if enableTemplate {
		projectRoot := os.Getenv("PROJECT_ROOT")

		/////////////////////// template[django]
		// https://ichi.pro/th/reiyn-ru-keiyw-kab-fiber-krxb-kar-phat-hna-web-golang-him-103546697181187
		// https://github.com/gofiber/template
		// django: https://github.com/gofiber/template/tree/master/django
		//engine := django.New("./web/django/views", ".django")
		engine := django.New(projectRoot+"./web/django/views", ".django")
		// register functions
		engine.AddFunc("nl2br", Nl2BrHtml)
		h.App = fiber.New(fiber.Config{Views: engine})
		return nil
	}

	h.App = fiber.New()
	return nil

}

// Nl2BrHtml Custom function for template
func Nl2BrHtml(value interface{}) string {
	if str, ok := value.(string); ok {
		return strings.Replace(str, "\n", "<br />", -1)
	}
	return ""
}

func (h *Handler) initFiberAppGroup() error {
	h.Api = h.App.Group("/api")         // /api
	h.Apiv1 = h.Api.Group("/v1")        // /api/v1
	h.Apiv2 = h.Api.Group("/v2")        // /api/v2
	h.Private = h.App.Group("/private") // /private
	return nil
}

func (h *Handler) initMiddleware() (err error) {
	// Internal middleware
	// middleware logger: https://github.com/gofiber/fiber/blob/master/middleware/logger/README.md
	h.App.Use(logger.New(logger.Config{
		Format:     "${blue}${time} ${pid} ${red}${status} ${method} ${yellow}${path} ${green}${ip} ${ua} ${bytesSent} ${white}${resBody}${reset}\n",
		TimeFormat: "2006/Jan/02",
		TimeZone:   "Asia/Bangkok",
	}))

	// Internal middleware: cache https://github.com/gofiber/fiber#-internal-middleware
	// curl -v http://127.0.0.1:3000/api/v2/greets
	// curl -v http://127.0.0.1:3000/api/v2/greets?refresh=true
	h.Apiv2.Use(cache.New(cache.Config{
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
	rsaKey, err := createRsaKey()
	if err != nil {
		return err
	}

	// set group private route to required jwt with rsa/key
	h.Private.Use(jwtware.New(jwtware.Config{
		SigningMethod: "RS256",
		SigningKey:    rsaKey.Public(),
	}))

	// keep key to buffer
	jwtCredSet(mykey, rsaKey)

	return nil
}

func (h *Handler) initRoute() (err error) {

	// GET plaintext
	// curl -v http://127.0.0.1:3000/
	h.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// GET Response plaintext with parameters: https://docs.gofiber.io/guide/routing#parameters
	// curl -v http://127.0.0.1:3000/user/foo/books/the_intelligent_investor?page=111
	h.App.Get("/user/:name/books/:title", func(c *fiber.Ctx) error {
		c.Set("Connection", "keep-alive")
		c.Set("Content-Type", "text/html; charset=utf-8")

		return c.SendString("name: " + c.Params("name") + " title: " + c.Params("title") + " page: " + c.Query("page"))
	})

	// GET Response JSON
	// curl -sv http://127.0.0.1:3000/health | jq
	h.App.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(&fiber.Map{
			"status": "ok",
		})
	})

	// GET Response JSON
	// curl -sv http://127.0.0.1:3000/api/v1/greets | jq
	h.Apiv1.Get("/greets", func(c *fiber.Ctx) error {
		var result []RespGreet
		for _, greet := range greets {
			result = append(result, greet)
		}

		log.Printf("Response json: %v\n", result)
		return c.Status(http.StatusOK).JSON(result)
	})

	// curl -sv http://127.0.0.1:3000/api/v2/greets | jq
	h.Apiv2.Get("/greets", func(c *fiber.Ctx) error {
		var result RespGreetV2
		for _, greet := range greets {
			result.Greets = append(result.Greets, greet)
		}

		result.UnixNano = time.Now().UnixNano()

		log.Printf("Response json: %v\n", result)
		return c.Status(http.StatusOK).JSON(result)
	})

	// Static files Server
	// curl http://127.0.0.1:3000/web
	h.App.Static("/web", "./web/static")

	// template[django]
	// curl http://127.0.0.1:3000/template/django/
	h.App.Get("/template/django/", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	// template[django]
	// curl http://127.0.0.1:3000/template/django/layout
	h.App.Get("/template/django/layout", func(c *fiber.Ctx) error {
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
	h.App.Post("/login", loginHandler)
	h.Private.Get("/", privateHandler)

	// Using Validator to validate the request struct
	// -[Fail case] curl -vs -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" --data "{\"name\":\"john\",\"isactive\":true}" http://localhost:3000/private/userinfo | jq
	// -[Success case] curl -vs -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" --data '{"tid":123456789,"name":"john","isactive":true,"email":"xxx@xxx.com","job":{"type":"salaryman","salary":10000}}' http://localhost:3000/private/userinfo | jq
	h.Private.Post("/userinfo", userInfoHandler)

	return nil
}
