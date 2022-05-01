package main_test

import (
	"github.com/Kong/go-pdk/test"
	"github.com/stretchr/testify/assert"
	"hello_world"
	"testing"
)

func TestShouldSetXPluginToHelloWhenMessageIsNotConfigured(t *testing.T) {
	// arrange
	conf := main.Config{}

	env, err := test.New(t, test.Request{
		Method:  "GET",
		Url:     "http://example.com",
		Headers: map[string][]string{"X-Hi": {"hello"}},
	})
	assert.NoError(t, err)

	// act
	env.DoHttps(&conf)

	// assert
	headerValue := env.ClientRes.Headers.Get("x-plugin")

	assert := assert.New(t)
	assert.Equal(headerValue, "hello", "should be equal")
}

func TestShouldSetXPluginToConfiguredMessage(t *testing.T) {
	// arrange
	conf := main.Config{
		Message: "test",
	}

	env, err := test.New(t, test.Request{
		Method:  "GET",
		Url:     "http://example.com",
		Headers: map[string][]string{"X-Hi": {"hello"}},
	})
	assert.NoError(t, err)

	// act
	env.DoHttps(&conf)

	// assert
	headerValue := env.ClientRes.Headers.Get("x-plugin")

	assert := assert.New(t)
	assert.Equal(headerValue, "test", "should be equal")
}
