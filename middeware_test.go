package plaud

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockReqMiddlewareBody struct {
	Type int `json:"type"`
}

func MockMiddleware(ctx *Context) *Error {
	body := MockReqMiddlewareBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		return ctx.AbortWithError("Could Not Decode Body", http.StatusInternalServerError)
	}
	if body.Type == 0 {
		return ctx.AbortWithError("test", http.StatusInternalServerError)
	}

	ctx.Next()
	return nil
}

func TestMiddlewareRoute(t *testing.T) {
	server := New(":8000")
	testRouter := NewRouter("/")
	testRouter.Post("/", func(ctx *Context) (*Data, *Error) {
		ctx.Status(http.StatusOK)
		_, err := ctx.Write([]byte("ok"))
		if err != nil {
			t.Log("[WARN] Could not write to response writer")
		}
		return nil, nil
	}).Use(MockMiddleware)
	server.Register(testRouter)

	body := bytes.NewBuffer([]byte(`{"type":1}`))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	res := httptest.NewRecorder()

	server.server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatal("Middleware did not let through")
	}

	if res.Body.String() != "ok" {
		t.Fatalf("Invalid Request Body")
	}

	body = bytes.NewBuffer([]byte(`{"type":0}`))
	req = httptest.NewRequest(http.MethodPost, "/", body)
	res = httptest.NewRecorder()

	server.server.ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("Middleware did not work correctly %d", res.Code)
	}
	mockError := &Error{}
	err := json.NewDecoder(res.Body).Decode(mockError)
	if err != nil {
		t.Fatal(err)
	}
	if mockError.Message != "test" {
		t.Fatalf("Invalid Request Body Got:%s", res.Body.String())
	}
}

func TestMiddlewareRouter(t *testing.T) {
	server := New(":8000")
	testRouter := NewRouter("/").Use(MockMiddleware)
	testRouter.Post("/ok", func(ctx *Context) (*Data, *Error) {
		ctx.Status(http.StatusOK)
		_, err := ctx.Write([]byte("ok"))
		if err != nil {
			t.Log("[WARN] Could not write to response writer")
		}
		return nil, nil
	})
	testRouter.Post("/not-ok", func(ctx *Context) (*Data, *Error) {
		ctx.Status(http.StatusOK)
		_, err := ctx.Write([]byte("ok"))
		if err != nil {
			t.Log("[WARN] Could not write to response writer")
		}
		return nil, nil
	})
	server.Register(testRouter)

	body := bytes.NewBuffer([]byte(`{"type":1}`))
	req := httptest.NewRequest(http.MethodPost, "/ok", body)
	res := httptest.NewRecorder()

	server.server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatal("Middleware did not let through")
	}

	if res.Body.String() != "ok" {
		t.Fatalf("Invalid Request Body")
	}

	body = bytes.NewBuffer([]byte(`{"type":0}`))
	req = httptest.NewRequest(http.MethodPost, "/not-ok", body)
	res = httptest.NewRecorder()

	server.server.ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("Middleware did not work correctly %d", res.Code)
	}
	mockError := &Error{}
	err := json.NewDecoder(res.Body).Decode(mockError)
	if err != nil {
		t.Fatal(err)
	}
	if mockError.Message != "test" {
		t.Fatalf("Invalid Request Body Got:%s", res.Body.String())
	}
}

func TestMiddleware(t *testing.T) {
	var results []string

	middleware := func(ctx *Context) *Error {
		results = append(results, "start")
		ctx.Next()
		results = append(results, "end")
		return nil
	}

	server := New(":8000")
	testRouter := NewRouter("/")
	testRouter.Get("/test", func(_ *Context) (*Data, *Error) {
		results = append(results, "executing")
		return nil, nil
	}).Use(middleware)

	server.Register(testRouter)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	res := httptest.NewRecorder()

	server.server.ServeHTTP(res, req)

	expected := []string{"start", "executing", "end"}

	for i, val := range results {
		if val != expected[i] {
			t.Fatalf("Middleware execution order is incorrect. Exepected: %v, got: %v", expected, results)
		}
	}
}

func TestMiddlewareAbort(t *testing.T) {
	var executionOrder []int8

	middleware1 := func(ctx *Context) *Error {
		executionOrder = append(executionOrder, 1)
		return ctx.AbortWithError("Testing testing 123", http.StatusInternalServerError)
	}

	middleware2 := func(_ *Context) *Error {
		executionOrder = append(executionOrder, 2)
		t.Fatal("This should get executed!")
		return nil
	}

	server := New(":8000")
	testRouter := NewRouter("/")

	testRouter.Get("/test", func(_ *Context) (*Data, *Error) {
		executionOrder = append(executionOrder, 3)
		t.Fatal("Handler function shouldn't get executed after any middleware returns error")
		return nil, nil
	}).Use(middleware1, middleware2)

	server.Register(testRouter)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	res := httptest.NewRecorder()

	server.server.ServeHTTP(res, req)

	if len(executionOrder) != 1 || executionOrder[0] != 1 {
		t.Fatal("Unexpected execution order of middlewares and handler function")
	}
}
