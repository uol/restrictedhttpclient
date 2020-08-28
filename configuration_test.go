package restrictedhttpclient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uol/funks"
	"github.com/uol/restrictedhttpclient"
)

// TestNullConfiguration - tests when a configuration is null
func TestNullConfiguration(t *testing.T) {
	var c *restrictedhttpclient.Configuration = nil
	_, err := restrictedhttpclient.New(c)
	assert.Equal(t, restrictedhttpclient.ErrNullConfiguration, err, "expected null configuration error")
}

// Test - tests a invalid number of max simultaneous requests
func TestInvalidMaxSimultaneousRequests(t *testing.T) {
	c := &restrictedhttpclient.Configuration{
		MaxSimultaneousRequests:   0,
		RequestTimeout:            *funks.ForceNewStringDuration("5s"),
		SkipCertificateValidation: true,
	}

	_, err := restrictedhttpclient.New(c)
	assert.Equal(t, restrictedhttpclient.ErrInvalidNumSimultaneousRequests, err, "expected invalid max number of requests error")
}

// Test - tests a invalid request timeout
func TestInvalidRequestTimeout(t *testing.T) {
	c := &restrictedhttpclient.Configuration{
		MaxSimultaneousRequests:   1,
		RequestTimeout:            *funks.ForceNewStringDuration("0s"),
		SkipCertificateValidation: true,
	}

	_, err := restrictedhttpclient.New(c)
	assert.Equal(t, restrictedhttpclient.ErrInvalidRequestTimeout, err, "expected invalid request timeout error")
}
