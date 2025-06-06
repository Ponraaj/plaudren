package plaud

// the type of middleware function
// the routers or the routes themself can have middlewares
// the middleware registered in the router takes precedence over the middleware registered in the routes
// if a error is returned the middleware chain is terminated , else the next middleware or the function is automatically called
type MiddleWareFunc func(c *Context) *Error
