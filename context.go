package plaud

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"
)

type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Middlewares    []MiddleWareFunc
	Errors         []*Error

	index int8
}

const abortIndex int8 = math.MaxInt8 >> 1

// wrapes the Request and ResponseWriter and returns a context instance
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		ResponseWriter: w,
		Request:        r,
		index:          -1,
		Errors:         make([]*Error, 0),
	}
}

func (c *Context) Status(code int) {
	c.ResponseWriter.WriteHeader(code)
}

func (c *Context) Header(key, value string) {
	c.ResponseWriter.Header().Add(key, value)
}

func (c *Context) Write(data []byte) (res int, err error) {
	res, err = c.ResponseWriter.Write(data)
	return res, err
}

func (c *Context) ErrorStack() string {
	if len(c.Errors) == 0 {
		return "No Errors"
	}

	var sb strings.Builder
	for _, err := range c.Errors {
		sb.WriteString(err.Error())
	}

	return sb.String()
}

// set the middlewares from route to context
func (c *Context) SetMiddlewares(middlewares []MiddleWareFunc) {
	c.Middlewares = middlewares
}

// Executes the chain of middlewares
// used only inside the middlewares
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.Middlewares)) {
		if c.Middlewares[c.index] != nil {
			if err := c.Middlewares[c.index](c); err != nil {
				c.Errors = append(c.Errors, err)
			}
		}
		c.index++
	}
}

// Aborts the middleware chain execution
func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

func (c *Context) AbortWithError(errMsg string, code int) *Error {
	err := NewError(errMsg).SetCode(code)
	c.Errors = append(c.Errors, err)

	c.Abort()
	return err
}

func (c *Context) BindJSON(obj any) *Error {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(obj); err != nil {
		return c.AbortWithError("Invalid JSON Format", http.StatusBadRequest)
	}
	return nil
}

func (c *Context) JSON(code int, obj any) {
	c.Header("Content-Type", "application/json")
	c.Status(code)
	if err := json.NewEncoder(c.ResponseWriter).Encode(obj); err != nil {
		c.Errors = append(c.Errors, NewError("Failed to encode JSON").SetCode(http.StatusInternalServerError))
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.Abort()
}
