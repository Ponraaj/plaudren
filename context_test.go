package plaud

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	ctx := NewContext(w, r)

	// check ResponseWriter
	if ctx.ResponseWriter != w {
		t.Fatal("ResponseWriter was not properly assigned to Context")
	}

	// check Request
	if ctx.Request != r {
		t.Fatal("Request was not properly assigned to Context")
	}

	middleware := func(ctx *Context) *Error {
		if ctx.Request.Method != http.MethodGet {
			t.Fatalf("Expected %v but got %v", http.MethodGet, ctx.Request.Method)
			return ctx.AbortWithError("Got wrong HTTP Method", http.StatusBadRequest)
		}
		ctx.Next()
		return nil
	}

	// check middleware func execution through context
	ctx.SetMiddlewares([]MiddleWareFunc{middleware})
	if len(ctx.Errors) > 0 {
		t.Fatal("Middleware failed")
	}

	errorMiddleware := func(ctx *Context) *Error {
		return ctx.AbortWithError("Test Error!!", http.StatusInternalServerError)
	}

	ctx.SetMiddlewares([]MiddleWareFunc{errorMiddleware})
	ctx.Next()
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected middleware error was not returned")
	}
}

func TestJSON(t *testing.T) {
	type Record struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	records := []Record{
		{
			Username: "Joseph Joestar",
			Password: "OhMyGod!!",
		},
		{
			Username: "Gyro Zeppeli",
			Password: "Pizza_Mozzarella",
		},
	}

	jsonBody, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("Failed to marshal json: %v", err)
	}

	server := New(":8000")
	testRouter := NewRouter("/")
	testRouter.Get("/test", func(ctx *Context) (*Data, *Error) {
		var data []Record
		if err := ctx.BindJSON(&data); err != nil {
			t.Fatalf("Failed to bind JSON: %v", err)
		}
		if data[0].Username != "Joseph Joestar" {
			t.Fatalf("Incorrect data parsed expected %v received %v", records[0].Username, data[0].Username)
		}

		if len(ctx.Errors) > 0 {
			t.Fatalf("Error returned by middleware: %v", ctx.ErrorStack())
		}
		ctx.JSON(http.StatusOK, data)
		return nil, nil
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", bytes.NewBuffer(jsonBody))
	r.Header.Set("Context-Type", "application/json")

	server.Register(testRouter)
	server.server.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatal("Failed to encode JSON")
	}

	var responseData []Record
	if err := json.Unmarshal(w.Body.Bytes(), &responseData); err != nil {
		t.Fatalf("Failed to Unmarshal JSON: %v", err)
	}

	if len(responseData) != len(records) {
		t.Fatalf("Expected %v records, got %v", len(records), len(responseData))
	}

	for i, record := range records {
		if responseData[i].Username != record.Username {
			t.Fatalf("Record %v username mismatch: expected %v, got %v",
				i, record.Username, responseData[i].Username)
		}
		if responseData[i].Password != record.Password {
			t.Fatalf("Record %v password mismatch: expected %v, got %v",
				i, record.Password, responseData[i].Password)
		}
	}
}
