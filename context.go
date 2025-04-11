package plaud

import "net/http"

type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Middlewares    []MiddleWareFunc
}

// wrapes the Request and ResponseWriter and returns a context instance
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		ResponseWriter: w,
		Request:        r,
	}
}

// set the middlewares from route to context
func (c *Context) SetMiddlewares(middlewares []MiddleWareFunc) {
	c.Middlewares = middlewares
}

// applies a set of all middlware to a route through context
func (c *Context) ApplyMiddlewares() *Error {
	for _, middleware := range c.Middlewares {
		if err := middleware(c.ResponseWriter, c.Request); err != nil {
			return err
		}
	}
	return nil
}
