package fiberserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.con/napat/gofiber/pkg/testHelper"
)

func setupApp() *fiber.App {
	app := fiber.New()

	rsaKey, _ := createRsaKey()
	jwtCredSet(mykey, rsaKey)

	app.Get("/", helloWorldHandler)
	app.Post("/login", loginHandler)

	return app
}

// go test -run TestHelloWorldHandler
// go test -v ./internal/fiberserver/
func TestHelloWorldHandler(t *testing.T) {

	t.Run("Success case: GET to / should return httpCode 200 with message: Hello, World!", func(t *testing.T) {

		app := setupApp()

		req := httptest.NewRequest("GET", "/", nil)
		//req.Header.Set("Authorization", bearerToken)

		resp, _ := app.Test(req)

		gotInt := resp.StatusCode
		wantInt := http.StatusOK
		testHelper.AssertCorrectHttpStatusCode(t, gotInt, wantInt)

		body, _ := ioutil.ReadAll(resp.Body)

		gotStr := string(body)
		wantStr := "Hello, World!"
		testHelper.AssertCorrectMessage(t, gotStr, wantStr)
	})

	t.Run("Fail case: POST to / should not accepted!(StatusMethodNotAllowed)", func(t *testing.T) {

		app := setupApp()

		req := httptest.NewRequest("POST", "/", nil)
		//req.Header.Set("Authorization", bearerToken)

		resp, _ := app.Test(req)

		gotInt := resp.StatusCode
		wantInt := http.StatusMethodNotAllowed
		testHelper.AssertCorrectHttpStatusCode(t, gotInt, wantInt)
	})
}

func TestMyPrinter(t *testing.T) {
	buffer := bytes.Buffer{}
	myPrinter(&buffer, "Foo")

	got := buffer.String()
	want := "MyPrinter: Foo"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// go test -run TestLoginHandler
// go test -v ./internal/fiberserver/
func TestLoginHandler(t *testing.T) {

	assertCorrectHttpStatusCode := func(t testing.TB, got, want int) {
		t.Helper()
		if got != want {
			t.Fatalf("got %v want %v", got, want)
		}
	}

	t.Run("Success case: POST to /login with john&pass=doe should be accept", func(t *testing.T) {

		app := setupApp()

		// FormValue
		reader := strings.NewReader("user=john&pass=doe")

		req, err := http.NewRequest(http.MethodPost, "/login", reader)
		if err != nil {
			t.Error(err)
		}

		//req.Header.Set("Authorization", bearerToken)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, _ := app.Test(req)

		gotInt := resp.StatusCode
		wantInt := http.StatusOK
		assertCorrectHttpStatusCode(t, gotInt, wantInt)

		resBody := RespLogin{}
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal([]byte(body), &resBody)
		if err != nil {
			t.Errorf("invalid response json data: got %v", []byte(body))
		}
		if len(resBody.Token) == 0 {
			t.Errorf("invalid jwt token: got len 0")
		}

		// check jwt token is valid
		// ...

	})
}
