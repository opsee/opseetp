package opseetp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testRequestValidator struct {
	Count int `json:"count"`
}

func (rv *testRequestValidator) Validate() error {
	if rv.Count == 1 {
		return nil
	}

	return fmt.Errorf("count is not 1")
}

func TestCORS(t *testing.T) {
	router := NewHTTPRouter(context.Background())
	corsDecoder := CORSRegexpDecodeFunc(
		[]string{"GET"},
		[]string{`http://(\w+\.)?(opsy\.co|opsee\.co|opsee)`},
	)

	router.Handle("GET", "/", []DecodeFunc{corsDecoder}, func(ctx context.Context) (interface{}, int, error) {
		return map[string]interface{}{"ok": true}, http.StatusOK, nil
	})

	req, err := http.NewRequest("GET", "http://potata.opsee/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "GET", rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "", rw.Header().Get("Access-Control-Allow-Origin"))

	req, err = http.NewRequest("GET", "http://potata.opsee/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://potata.opsee")

	rw = httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "http://potata.opsee", rw.Header().Get("Access-Control-Allow-Origin"))
}

func TestAuthorization(t *testing.T) {
	userKey := 0

	router := NewHTTPRouter(context.Background())
	authDecoder := AuthorizationDecodeFunc(userKey, user{})

	router.Handle("GET", "/", []DecodeFunc{authDecoder}, func(ctx context.Context) (interface{}, int, error) {
		return ctx.Value(userKey), http.StatusOK, nil
	})

	req, err := http.NewRequest("GET", "http://potata.opsee/", nil)
	if err != nil {
		t.Fatal(err)
	}

	token := base64.StdEncoding.EncodeToString([]byte(`{"id": 1, "email": "cliff@leaninto.it"}`))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", token))

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)

	u := &user{}
	err = json.NewDecoder(rw.Body).Decode(u)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, u.Id, 1)
	assert.Equal(t, u.Email, "cliff@leaninto.it")
}

func TestRequest(t *testing.T) {
	requestKey := 0

	router := NewHTTPRouter(context.Background())
	requestDecoder := RequestDecodeFunc(requestKey, testRequestValidator{})

	router.Handle("GET", "/", []DecodeFunc{requestDecoder}, func(ctx context.Context) (interface{}, int, error) {
		return ctx.Value(requestKey), http.StatusOK, nil
	})

	req, err := http.NewRequest("GET", "http://potata.opsee/", bytes.NewBufferString(`{"count": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)

	rv := &testRequestValidator{}
	err = json.NewDecoder(rw.Body).Decode(rv)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, rv.Count)

	req, err = http.NewRequest("GET", "http://potata.opsee/", bytes.NewBufferString(`{"count": 2}`))
	if err != nil {
		t.Fatal(err)
	}

	rw = httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}
