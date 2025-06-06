package plaud

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// type of the function that handles the http request
type HTTPFunc func(*Context) (*Data, *Error)

// type of the router which registers the functions
type HTTPRouter interface {
	// http methods
	// the param func should implement the HTTPFunc interface
	Get(string, HTTPFunc) HTTPRoute
	Post(string, HTTPFunc) HTTPRoute
	Put(string, HTTPFunc) HTTPRoute
	Patch(string, HTTPFunc) HTTPRoute
	Delete(string, HTTPFunc) HTTPRoute

	// static files
	// serves the contents of the directory
	ServeDir(string, http.FileSystem) HTTPRoute

	// returns all the routes
	GetRoutes() []HTTPRoute

	// returns all the fileHandlers
	GetHandlers() []HTTPRoute

	// register another router with the current router
	Handle(string, HTTPRouter)

	// register the router with a mux to handle http transport
	RegisterServer(*http.ServeMux)

	// called before the router is attached to the server
	Register()

	// registers a set of middlewares for the routers
	// applied to every route registered within the router
	// takes precedence over the middleware within the route
	Use(...MiddleWareFunc) HTTPRouter
}

// router contains a group of routes
// should implement the HTTPRouter interface
type Router struct {
	path   string
	routes []HTTPRoute

	fileHandlers []HTTPRoute
	middlewares  []MiddleWareFunc
}

func NewRouter(path string) *Router {
	return &Router{
		path: strings.TrimRight(path, "/"),
	}
}

func (r *Router) createRoute(method HTTPMethod, path string, httpFunc HTTPFunc) HTTPRoute {
	path = strings.TrimRight(path, "/")
	route, err := NewRoute(method, r.path+path, httpFunc)
	if err != nil {
		slog.Error("Invalid route", "path", path)
		return nil
	}
	r.routes = append(r.routes, route)

	return route
}

func (r *Router) Get(path string, httpFunc HTTPFunc) HTTPRoute {
	return r.createRoute(GET, path, httpFunc)
}

func (r *Router) Put(path string, httpFunc HTTPFunc) HTTPRoute {
	return r.createRoute(PUT, path, httpFunc)
}

func (r *Router) Post(path string, httpFunc HTTPFunc) HTTPRoute {
	return r.createRoute(POST, path, httpFunc)
}

func (r *Router) Patch(path string, httpFunc HTTPFunc) HTTPRoute {
	return r.createRoute(PATCH, path, httpFunc)
}

func (r *Router) Delete(path string, httpFunc HTTPFunc) HTTPRoute {
	return r.createRoute(DELETE, path, httpFunc)
}

func (r *Router) ServeDir(path string, dir http.FileSystem) HTTPRoute {
	path = strings.TrimRight(path, "/")
	handler := NewFileHandler(dir, r.path+path)
	r.fileHandlers = append(r.fileHandlers, handler)
	return handler
}

func (r *Router) GetRoutes() []HTTPRoute {
	return r.routes
}

func (r *Router) GetHandlers() []HTTPRoute {
	return r.fileHandlers
}

// Registers a router with the given path
func (r *Router) Handle(path string, router HTTPRouter) {
	// calls the initialization of the router
	router.Register()

	path = strings.TrimRight(path, "/")

	r.path = strings.TrimRight(r.path, "/")
	r.path = fmt.Sprintf("%s%s", r.path, path)

	for _, route := range router.GetRoutes() {
		route.Prepend(r.path)
		route.stackMiddleware(r.middlewares)
		r.routes = append(r.routes, route)
	}

	for _, route := range router.GetHandlers() {
		route.Prepend(r.path)
		route.stackMiddleware(r.middlewares)
		r.fileHandlers = append(r.fileHandlers, route)
	}
}

// empty function i dont know why but should be there to implement the HTTPRouter interface
func (r *Router) Register() {
}

func (r *Router) RegisterServer(mux *http.ServeMux) {
	for _, route := range r.routes {
		slog.Info("Api Route", "route", route.GetRoute())
		route.stackMiddleware(r.middlewares)
		mux.HandleFunc(route.GetRoute(), route.GetHandleFunc())
	}
	for _, handler := range r.fileHandlers {
		handler.stackMiddleware(r.middlewares)

		// handle routes without trailing
		// TODO: danger ahead
		// if handler.GetRoute() != "/" {
		// 	handlePath:=strings.TrimRight(handler.GetRoute(), "/")
		// 	slog.Info("Registered", "path", handlePath)
		// 	mux.HandleFunc(handlePath, func(w http.ResponseWriter, r *http.Request) {
		// 		http.StripPrefix(handlePath, handler.GetHandler()).ServeHTTP(w, r)
		// 	})
		// }

		mux.Handle(handler.GetRoute(), http.StripPrefix(handler.GetRoute(), handler.GetHandler()))
	}
}

// middleware stuff
// registers the middleware for entire router
func (r *Router) Use(middlewares ...MiddleWareFunc) HTTPRouter {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}
