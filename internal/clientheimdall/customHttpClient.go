package clientheimdall

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type MyHTTPClient struct {
	Client http.Client
}

func (c *MyHTTPClient) Do(request *http.Request) (*http.Response, error) {
	// Force add header to all requests that use the MyHTTPClient
	request.SetBasicAuth("username", "passwd")
	request.Header.Set("X-Sample-Header", "my token")

	// Access Request Body to write signature header
	if request.Body != nil {
		//https://stackoverflow.com/a/46948073/3616311
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		request.Body.Close() //  must close
		request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		nowString := time.Now().String()
		request.Header.Set("X-Test-Time", nowString)
		request.Header.Set("X-Test-Signature", calSignature(bodyBytes, nowString))

		fmt.Printf("++++++++++++++++\n"+
			"%s\n"+
			"++++++++++++\n",
			string(bodyBytes))
	}

	return c.Client.Do(request)
}

// calSignature dummy function, only for demo
func calSignature(requestBodyString []byte, iv string) string {
	return fmt.Sprintf("%s:%d", iv, len(requestBodyString))
}
