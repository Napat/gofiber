package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.con/napat/gofiber/internal/clientheimdall"

	"github.com/gojektech/heimdall/v6"
	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/gojektech/heimdall/v6/hystrix"
)

func main() {
	// fmt.Println("Hello")
	// getGoogle()
	// fmt.Println("----------------------")
	// getWithRetry()
	// fmt.Println("----------------------")
	// customHttpClient()
	// fmt.Println("----------------------")
	// useCustomPlugpin()
	// fmt.Println("----------------------")
	// postWithMySignature()
	fmt.Println("----------------------")
	wg := &sync.WaitGroup{}
	for id := 0; id < 10; id++ {
		wg.Add(1)
		go hystrixClientCircuitBreaker(wg, id)
	}
	wg.Wait()
}

func getGoogle() {
	// Create a new HTTP client with a default timeout
	timeout := 1000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// Use the clients GET method to create and execute the request
	resp, err := client.Get("http://google.com", nil)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Heimdall returns the standard *http.Response object
	body, err := ioutil.ReadAll(resp.Body) // Deprecation of ioutil: https://golang.org/doc/go1.16#ioutil
	fmt.Println(string(body))
}

func getWithRetry() {

	retryCount := 4 // maximum send = 1(first) + 4(retry) = 5 times

	/// Set retry mechanism
	// Constant Backoff mechanism. Constant backoff increases the backoff at a constant rate
	//backoffInterval := 2 * time.Millisecond
	backoffInterval := 1 * time.Second

	// Define a maximum jitter interval. It must be more than 1*time.Millisecond
	maximumJitterInterval := 5 * time.Millisecond

	backoff := heimdall.NewConstantBackoff(backoffInterval, maximumJitterInterval)

	// Create a new retry mechanism with the backoff
	retrier := heimdall.NewRetrier(backoff)

	timeout := 1000 * time.Millisecond
	// Create a new client, sets the retry mechanism, and the number of times you would like to retry
	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(retryCount),
	)

	/// request GET
	// Create an http.Request instance
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:3000/api/v2/randomnotresponse?refresh=true", nil)
	if err != nil {
		panic(err)
	}

	// Call the `Do` method, which has a similar interface to the `http.Do` method
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body) // Deprecation of ioutil: https://golang.org/doc/go1.16#ioutil
	fmt.Println(string(body))
}

func customHttpClient() {
	client := httpclient.NewClient(
		httpclient.WithHTTPClient(&clientheimdall.MyHTTPClient{
			Client: http.Client{Timeout: 25 * time.Millisecond},
		}),
	)

	/// request GET
	// Create an http.Request instance
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:3000/api/v2/randomnotresponse?refresh=true", nil)
	if err != nil {
		panic(err)
	}

	// Call the `Do` method, which has a similar interface to the `http.Do` method
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body) // Deprecation of ioutil: https://golang.org/doc/go1.16#ioutil
	fmt.Println(string(body))
}

func useCustomPlugpin() {
	// Create a new HTTP client with a default timeout
	timeout := 1000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))
	// client := heimdall.NewHTTPClient(timeout)

	requestLogger := clientheimdall.NewRequestLogger(nil, nil)
	client.AddPlugin(requestLogger)

	req, _ := http.NewRequest(http.MethodGet, "http://google.com", nil)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("+++++")
	fmt.Println("res: ", resp)
}

// postWithMySignature ...
func postWithMySignature() {
	client := httpclient.NewClient(
		httpclient.WithHTTPClient(&clientheimdall.MyHTTPClient{
			Client: http.Client{Timeout: 250 * time.Millisecond},
		}),
	)

	requestBodyString := `{"firstname":"Foo","lastname":"Bar"}`
	requestBody := bytes.NewReader([]byte(requestBodyString))

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept-Language", "en")

	// Only for demo
	// Note: The better way to prevent duplicate signature header code is to use
	// customHttpClient(see customHttpClient.go)
	nowString := time.Now().String()
	headers.Set("X-My-Time", nowString)
	headers.Set("X-My-Signature", calSignature([]byte(requestBodyString), nowString))

	resp, err := client.Post("http://127.0.0.1:3000/api/v2/test_signature?refresh=true", requestBody, headers)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body) // Deprecation of ioutil: https://golang.org/doc/go1.16#ioutil
	if err != nil {
		panic(err)
	}
	fmt.Println(string(respBody))
}

// calSignature dummy function, only for demo
func calSignature(requestBodyString []byte, iv string) string {
	return fmt.Sprintf("%s:%d", iv, len(requestBodyString))
}

func hystrixClientCircuitBreaker(wg *sync.WaitGroup, id int) {

	defer wg.Done()

	timeout := 10 * time.Second

	// Create a new fallback function
	// fallback is called when all retry are error
	// return error of this fallback will return to _, ***err*** := client.Do()
	fallbackFn := func(err error) error {
		log.Printf("hystrix fallbackFN with error: %v\n", err)

		//_, err := http.Post("post_to_channel_two")
		// err = nil

		return err
	}

	retrier := retrerConfig()
	retryCount := 4 // maximum send = 1(first) + 4(retry) = 5 times

	cmdName := fmt.Sprintf("get_ccbreaker_%v", id)

	// Create a new hystrix-wrapped HTTP client with the command name, along with other required options
	client := hystrix.NewClient(
		hystrix.WithHTTPTimeout(timeout),
		hystrix.WithRetrier(retrier),
		hystrix.WithRetryCount(retryCount),
		hystrix.WithCommandName(cmdName),
		hystrix.WithHystrixTimeout(timeout), // how long to wait for command to complete, in time.Duration
		hystrix.WithMaxConcurrentRequests(100),
		hystrix.WithErrorPercentThreshold(25), // causes circuits to open once the rolling measure of errors exceeds this percent of requests
		hystrix.WithSleepWindow(5000),         // how long, in milliseconds, to wait after a circuit opens before testing for recovery
		hystrix.WithRequestVolumeThreshold(5), // the minimum number of requests needed before a circuit can be tripped due to health
		hystrix.WithFallbackFunc(fallbackFn),
		// hystrix.WithStatsDCollector("localhost:8125", "myapp.hystrix"),
	)

	/// request GET
	// Create an http.Request instance
	// req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:3000/api/v2/ccbreaker_respond?refresh=true", nil)
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:3000/api/v1/ccbreaker_respond", nil)
	if err != nil {
		panic(err)
	}

	// Call the `Do` method, which has a similar interface to the `http.Do` method
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("respond error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body) // Deprecation of ioutil: https://golang.org/doc/go1.16#ioutil
	fmt.Println(string(body))

}

func retrerConfig() heimdall.Retriable {
	/// Set retry mechanism
	// Constant Backoff mechanism. Constant backoff increases the backoff at a constant rate
	//backoffInterval := 2 * time.Millisecond
	backoffInterval := 1 * time.Second

	// Define a maximum jitter interval. It must be more than 1*time.Millisecond
	maximumJitterInterval := 5 * time.Millisecond

	backoff := heimdall.NewConstantBackoff(backoffInterval, maximumJitterInterval)

	// Create a new retry mechanism with the backoff
	retrier := heimdall.NewRetrier(backoff)
	return retrier
}
