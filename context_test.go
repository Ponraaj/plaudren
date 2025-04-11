package plaud

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := NewContext(w, r)

	// check ResponseWriter
	if ctx.ResponseWriter != w {
		t.Fatal("ResponseWriter was not properly assigned to Context")
	}

	// check Request
	if ctx.Request != r {
		t.Fatal("Request was not properly assigned to Context")
	}

	var middleware MiddleWareFunc = func(w http.ResponseWriter, r *http.Request) *Error {
		return nil
	}

	// check middleware func execution through context
	ctx.SetMiddlewares([]MiddleWareFunc{middleware})
	if err := ctx.ApplyMiddlewares(); err != nil {
		t.Fatal("Middleware failed")
	}
}
