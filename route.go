package plaud

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type HTTPMethod string

// allowed methods
const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	PATCH  HTTPMethod = "PATCH"
	DELETE HTTPMethod = "DELETE"
)

// leaf node of a router
type HTTPRoute interface {
	GetRoute() string
	GetHandleFunc() func(http.ResponseWriter, *http.Request)
	GetHandler() http.Handler

	// registers a pre route from a router or handler
	Prepend(string)

	// registers the routers middlewares to the route
	stackMiddleware([]MiddleWareFunc)
	// registers route specific middleware
	Use(...MiddleWareFunc)
}

// handles the URI of the api which is to be registered with the Router
type Route struct {
	method      HTTPMethod
	path        string
	httpfunc    HTTPFunc
	middlewares []MiddleWareFunc
}

func (route *Route) GetRoute() string {
	return fmt.Sprintf("%s %s", route.method, route.path)
}

// return the http handler for the routes
// handles the encoding (json,grpc...)
//
// TODO: Error handling will be added in a future commit
func (route *Route) GetHandleFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)

		handlers := make([]MiddleWareFunc, len(route.middlewares)+1)
		copy(handlers, route.middlewares)

		handlers[len(route.middlewares)] = func(ctx *Context) *Error {
			data, err := route.httpfunc(ctx)
			if err != nil {
				ctx.JSON(err.code, err)
				return err
			}

			if data != nil {
				ctx.JSON(data.code, data)
			}

			return nil
		}
		ctx.SetMiddlewares(handlers)
		// handling middlewares
		ctx.Next()
		if len(ctx.Errors) > 0 {
			err := ctx.Errors[len(ctx.Errors)-1]
			// TODO: handle error below
			// have a default logger with the router
			ctx.JSON(err.code, err)
		}
	}
}

func (route *Route) GetHandler() http.Handler {
	return nil
}

func NewRoute(method HTTPMethod, path string, httpfunc HTTPFunc) (*Route, error) {
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		return nil, errors.New("invalid route")
	}
	return &Route{
		method:   method,
		path:     path,
		httpfunc: httpfunc,
	}, nil
}

func (route *Route) stackMiddleware(middleware []MiddleWareFunc) {
	route.middlewares = append(middleware, route.middlewares...)
}

// registers a set of all middlewares
// adds the middlewares in order
func (route *Route) Use(middlewares ...MiddleWareFunc) {
	route.middlewares = append(route.middlewares, middlewares...)
}

func (route *Route) Prepend(path string) {
	route.path = fmt.Sprintf("%s%s", path, strings.TrimRight(route.path, "/"))
}
