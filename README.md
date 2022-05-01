# Kong go plugin guide

This repo is a complete up to date guide to creating, testing, installing, and running a go plugin for the kong api gateway.

I am creating this repo to document my notes on creating a go plugin because I found the [official documentation](https://docs.konghq.com/gateway/2.8.x/reference/external-plugins/) difficult to follow and most [existing tutorials](https://konghq.com/blog/kong-gateway-go-plugin) are out of date. Hopefully this will save you a few hours of experimentation.

If you find a mistake, or would like to contribute an additional example plugin please open a PR.

# Plugin overview

Kong natively supports plugins created in the lua scripting language, these are executed directly by the kong gateway. In 2019 support was added to create plugins in other languages, currently `go`, `python`, and `js` are supported.

These 'other language' plugins integrate with kong by running a local sever that provides request lifecycle hooks.

> The server for a `go` plugin is embedded in the [`go-pdk`](https://github.com/Kong/go-pdk) package.
> Most existing tutorials are using the deprecated [`go-pluginserver`](https://github.com/Kong/go-pluginserver), do not use this.

## Plugin skeleton

This a a complete example plugin that appends a response header of `x-plugin-header` to each request. This plugin has an optional configuration option `message` that can be used to configure the header value.

```go
package hello_world_plugin

import (
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
)

func main() {
	server.StartServer(New, Version, Priority)
}

var Version = "0.2"
var Priority = 5000

type Config struct {
  Message string
}

func New() interface{} {
	return &Config{}
}

func (conf Config) Access(kong *pdk.PDK) {
  message := conf.Message
	if message == "" {
		message = "Hello world!"
	}
	kong.Response.SetHeader("x-plugin-header", message)
}
```

## Configuration

To add configuration settings, simply add properties to the `Config` struct. These properties can be referenced in the declarative config file in lowercase.

`main.go`
```go
type Config struct {
  Message string
}
```

`declarative_config.yaml`
```yaml
plugins:
- name: hello_world
  config:
    message: example
```

## Priority

This is an important property that dictates the [order in which your plugin will execute](https://docs.konghq.com/gateway/2.8.x/plugin-development/custom-logic/#plugins-execution-order). Plugins will execute in order of highest to lowest.

| Plugin 	| Priority |
|---|---|
| pre-function 	+| inf |
| correlation-id | 100001 |
| zipkin 	| 100000 |
| bot-detection 	| 2500 |
| cors 	| 2000 |
| session 	| 1900 |
| jwt 	| 1005 |
| oauth2 	| 1004 |
| key-auth 	| 1003 |
| ldap-auth 	| 1002 |
| basic-auth 	| 1001 |
| hmac-auth 	| 1000 |
| grpc-gateway 	| 998 |
| ip-restriction 	| 990 |
| request-size-limiting 	| 951 |
| acl 	| 950 |
| rate-limiting 	| 901 |
| response-ratelimiting 	| 900 |
| request-transformer 	| 801 |
| response-transformer 	| 800 |
| aws-lambda 	| 750 |
| azure-functions 	| 749 |
| prometheus 	| 13 |
| http-log 	| 12 |
| statsd 	| 11 |
| datadog 	| 10 |
| file-log 	| 9 |
| udp-log 	| 8 |
| tcp-log 	| 7 |
| loggly 	| 6 |
| syslog 	| 4 |
| grpc-web 	| 3 |
| request-termination 	| 2 |
| correlation-id | 1 |
| post-function | -1000 |

## Phases

These are the request lifecycle events that you plugin can interact with. Official description here: https://docs.konghq.com/gateway/2.8.x/plugin-development/custom-logic/#available-contexts

| phase | description |
|---|---|
| Certificate | Executed during the SSL certificate serving phase of the SSL handshake. |
| Rewrite | Executed for every request upon its reception from a client as a rewrite phase handler.
In this phase, neither the Service nor the Consumer have been identified, hence this handler will only be executed if the plugin was configured as a global plugin. |
| Access | Executed for every request from a client and before it is being proxied to the upstream service. |
| Response | Replaces both header_filter() and body_filter(). Executed after the whole response has been received from the upstream service, but before sending any part of it to the client. |
| Preread | For stream connections - Executed once for every connection. |
| Log | Executed when the last response byte has been sent to the client. |

Each phase can be intercepted by exporting a simple go function with the matching phase name

```go
func (conf Config) Rewrite(kong *pdk.PDK) {
}

func (conf Config) Access(kong *pdk.PDK) {
}

func (conf Config) Log(kong *pdk.PDK) {
}
```

# Testing

Run tests with `go test`

[A test harness is provided by the `go-pdk` library.](https://github.com/Kong/go-pdk/blob/master/test/test.go#L53).

The following is taken from the source file:

Trivial example:
```go
	package main
	import (
		"testing"
		"github.com/Kong/go-pdk/test"
		"github.com/stretchr/testify/assert"
	)
	func TestPlugin(t *testing.T) {
		// arrange
		env, err := test.New(t, test.Request{
			Method: "GET",
			Url:    "http://example.com?q=search&x=9",
			Headers: map[string][]string{ "X-Hi": {"hello"}, },
		})
		assert.NoError(t, err)

		// act
		env.DoHttps(&Config{})

		// assert
		assert.Equal(t, 200, env.ClientRes.Status)
		assert.Equal(t, "Go says Hi!", env.ClientRes.Headers.Get("x-hello-from-go"))
	}
  ```
>in short:
>1. Create a test environment passing a test.Request{} object to the test.New() function.
>2. Create a Config{} object (or the appropriate config structure of the plugin)
>3. env.DoHttps(t, &config) will pass the request object through the plugin, exercising
each event handler and return (if there's no hard error) a simulated response object.
There are other env.DoXXX(t, &config) functions for HTTP, TCP, TLS and individual phases.
3.5 The http and https functions assume the service response will be an "echo" of the
request (same body and headers) if you need a different service response, use the
individual phase methods and set the env.ServiceRes object manually.
>4. Do assertions to verify the service request and client response are as expected.

## Testing limitations

- Only `GET` is supported in the testing struct

```go
// Validate verifies a request and normalizes the headers.
// (to make them case-insensitive)
func (req *Request) Validate() error {
	_, err := url.Parse(req.Url)
	if err != nil {
		return err
	}

	req.Headers = mergeHeaders(make(http.Header), req.Headers)

	if req.Method == "GET" {
		if req.Body != "" {
			return fmt.Errorf("GET requests must not have body, found \"%v\"", req.Body)
		}
		return nil
	}
	return fmt.Errorf("Unsupported method \"%v\"", req.Method)
}
```

# Installing

To install the plugin build the binary and update the the kong.conf file.

build

```sh
go build -o hello_world .
```

kong.conf

https://docs.konghq.com/gateway/2.7.x/reference/external-plugins/#kong-gateway-plugin-server-configuration

```conf
#add your plugin name to the installed plugins list
plugins=bundled,hello_world

#since this is not a lua plugin, you must provide the name of the plugin server. This is a csv list so multiple plugins can be added here.
pluginserver_names=hello_world

#inform kong how to start and query the plugin
pluginserver_hello_world_start_cmd=/kong/go-plugins/hello_world
pluginserver_hello_world_query_cmd=/kong/go-plugins/hello_world -dump
```

# Running the example

```sh
docker-compose build && docker-compose up
```
