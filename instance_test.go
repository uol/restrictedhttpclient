package restrictedhttpclient_test

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	gtutils "github.com/uol/gotest/utils"

	gthttp "github.com/uol/gotest/http"

	"github.com/stretchr/testify/assert"
	"github.com/uol/funks"
	"github.com/uol/restrictedhttpclient"
)

func createClient(numRequests int) *restrictedhttpclient.Instance {

	c := &restrictedhttpclient.Configuration{
		MaxSimultaneousRequests:   uint64(numRequests),
		RequestTimeout:            *funks.ForceNewStringDuration("5s"),
		SkipCertificateValidation: true,
	}

	instance, err := restrictedhttpclient.New(c)
	if err != nil {
		panic(err)
	}

	return instance
}

const (
	testServerHost string = "localhost"
	testServerPort int    = 18080
	testEndpoint   string = "/test"
)

var testURL string = fmt.Sprintf("http://%s:%d%s", testServerHost, testServerPort, testEndpoint)

// createHTTPServer - creates a new test server
func createHTTPServer() *gthttp.Server {

	serverConf := &gthttp.Configuration{
		Host:        testServerHost,
		Port:        testServerPort,
		ChannelSize: 100,
		Responses: map[string][]gthttp.ResponseData{
			"normal": {
				{
					RequestData: gthttp.RequestData{
						Method: "GET",
						URI:    testEndpoint,
					},
					Status: http.StatusOK,
					Wait:   funks.ForceNewStringDuration("2s").Duration,
				},
			},
		},
	}

	return gthttp.NewServer(serverConf)
}

// TestRestrictionNotReached - tests a random number of simultaneous request, no error
func TestRestrictionNotReached(t *testing.T) {

	maxReqs := gtutils.RandomInt(3, 6)
	c := createClient(maxReqs)
	s := createHTTPServer()
	defer s.Close()
	wg := sync.WaitGroup{}

	var numSuccess uint32

	for i := 0; i < maxReqs; i++ {
		wg.Add(1)
		go func() {
			r, err := c.Get(testURL)

			if !assert.NoError(t, err, "expected no errors: ", err) {
				wg.Done()
				return
			}

			if !assert.Equal(t, http.StatusOK, r.StatusCode, "expected 200 as status code") {
				wg.Done()
				return
			}

			atomic.AddUint32(&numSuccess, 1)
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, maxReqs, int(numSuccess), "the number of successes not match to the number of requests")
}

func testRequestRestrictions(t *testing.T, c *restrictedhttpclient.Instance, numReqs, numExpectedSuccess, numExpectedErrors int) {

	wg := sync.WaitGroup{}

	var numSuccess, numErrors uint32

	for i := 0; i < numReqs; i++ {
		wg.Add(1)
		go func() {
			r, err := c.Get(testURL)

			if err != nil {
				if assert.Equal(t, restrictedhttpclient.ErrMaxRequestsReached, err, "unexpected error type: ", err) {
					atomic.AddUint32(&numErrors, 1)
				}
				wg.Done()
				return
			}

			if !assert.Equal(t, http.StatusOK, r.StatusCode, "expected 200 as status code") {
				wg.Done()
				return
			}

			atomic.AddUint32(&numSuccess, 1)
			wg.Done()
		}()
	}

	wg.Wait()

	if !assert.Equal(t, numExpectedSuccess, int(numSuccess), "the number of successes does not match to the number of requests") {
		return
	}

	assert.Equal(t, numExpectedErrors, int(numErrors), "the number of errors does not match to the number of expected errors")
}

// TestRestrictionReached - tests a random number of simultaneous request, with errors
func TestRestrictionReached(t *testing.T) {

	maxReqs := gtutils.RandomInt(3, 6)
	expectedErrors := gtutils.RandomInt(3, 6)
	c := createClient(maxReqs)
	s := createHTTPServer()
	defer s.Close()

	testRequestRestrictions(t, c, maxReqs+expectedErrors, maxReqs, expectedErrors)

}

// TestRequestBursts - tests some requests bursts
func TestRequestBursts(t *testing.T) {

	maxReqs := gtutils.RandomInt(2, 5)
	bursts := gtutils.RandomInt(3, 6)
	c := createClient(maxReqs)
	s := createHTTPServer()
	defer s.Close()

	for b := 0; b < bursts; b++ {

		numReqs := gtutils.RandomInt(10, 20)

		testRequestRestrictions(t, c, numReqs, maxReqs, numReqs-maxReqs)
	}
}
