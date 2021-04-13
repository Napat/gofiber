package fiberserver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.con/napat/gofiber/pkg/testHelper"
)

// func TestGetFilename(t *testing.T) {
// 	_, filename, _, _ := runtime.Caller(0)
// 	t.Logf("Current test filename: %s", filename)
// }

// go test -run TestPrivateHandler
// export PROJECT_ROOT=$PWD/ & go test -v ./internal/fiberserver/
func TestPrivateHandler(t *testing.T) {

	t.Run("Success case: GET to / should return httpCode 200 with message: Welcome john", func(t *testing.T) {

		// assertCorrectMessage := func(t testing.TB, got, want string) {
		// 	t.Helper()
		// 	if got != want {
		// 		t.Fatalf("got %q want %q", got, want)
		// 	}
		// }

		// assertCorrectHttpStatusCode := func(t testing.TB, got, want int) {
		// 	t.Helper()
		// 	if got != want {
		// 		t.Fatalf("got %v want %v", got, want)
		// 	}
		// }

		// assertNoError := func(t testing.TB, err error) {
		// 	if err != nil {
		// 		_, file, line, _ := runtime.Caller(1) // similar with t.Helper()
		// 		t.Fatalf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		// 	}
		// }

		serv := NewServHandler()

		// Request Login
		// FormValue
		reader := strings.NewReader("user=john&pass=doe")

		req, err := http.NewRequest(http.MethodPost, "/login", reader)
		testHelper.AssertJustError(t, err)

		//req.Header.Set("Authorization", bearerToken)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, _ := serv.App.Test(req)

		// Check http status code
		gotInt := resp.StatusCode
		wantInt := http.StatusOK
		testHelper.AssertCorrectHttpStatusCode(t, gotInt, wantInt)

		resBody := RespLogin{}
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal([]byte(body), &resBody)
		if err != nil {
			t.Errorf("invalid response json data: got %v", []byte(body))
		}
		if len(resBody.Token) == 0 {
			t.Errorf("invalid jwt token: got len 0")
		}

		bearerToken := "Bearer " + resBody.Token

		// Request /private/
		// with header "Authorization: Bearer $TOKEN"
		req = httptest.NewRequest("GET", "/private", nil)
		req.Header.Set("Authorization", bearerToken)

		resp, _ = serv.App.Test(req)

		// Check http status code
		gotInt = resp.StatusCode
		wantInt = http.StatusOK
		testHelper.AssertCorrectHttpStatusCode(t, gotInt, wantInt)

		body, _ = ioutil.ReadAll(resp.Body)

		gotStr := string(body)
		wantStr := "Welcome John Doe"
		testHelper.AssertCorrectMessage(t, gotStr, wantStr)

	})
}
