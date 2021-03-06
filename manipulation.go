package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type Constructor struct {
	request *http.Request
	payload Payload
}

func NewConstructor(req *http.Request, payload Payload) Constructor {
	c := Constructor{request: req, payload: payload}
	return c
}

func (c *Constructor) ApplyMiddleware(middleware string) error {

	newPayload, err := ExecuteMiddleware(middleware, c.payload)

	if err != nil {
		log.WithFields(log.Fields{
			"error":      err.Error(),
			"middleware": middleware,
		}).Error("Error during middleware transformation, not modifying payload!")

		return err
	} else {

		log.WithFields(log.Fields{
			"middleware": middleware,
		}).Info("Middleware transformation complete!")
		// override payload with transformed new payload
		c.payload = newPayload

		return nil
	}
}

// reconstructResponse changes original response with details provided in Constructor Payload.Response
func (c *Constructor) reconstructResponse() *http.Response {
	response := &http.Response{}
	response.Request = c.request

	// adding headers
	response.Header = make(http.Header)

	// applying payload
	if len(c.payload.Response.Headers) > 0 {
		for k, values := range c.payload.Response.Headers {
			// headers is a map, appending each value
			for _, v := range values {
				response.Header.Add(k, v)
			}

		}
	}
	// adding body, length, status code
	buf := bytes.NewBufferString(c.payload.Response.Body)
	response.ContentLength = int64(buf.Len())
	response.Body = ioutil.NopCloser(buf)
	response.StatusCode = c.payload.Response.Status

	return response
}

// reconstructRequest changes original request with details provided in Constructor Payload.Request
func (c *Constructor) reconstructRequest() *http.Request {
	request := c.request

	request.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(c.payload.Request.Body)))
	request.RequestURI = ""
	request.Host = c.payload.Request.Destination
	request.Method = c.payload.Request.Method
	request.URL.Path = c.payload.Request.Path
	request.URL.RawQuery = c.payload.Request.Query
	request.RemoteAddr = c.payload.Request.RemoteAddr
	request.Header = c.payload.Request.Headers

	return request
}
