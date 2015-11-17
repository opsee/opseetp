package opseetp

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"reflect"
)

type RequestValidator interface {
	Validate() error
}

func RequestDecodeFunc(requestKey int, requestType interface{}) DecodeFunc {
	return func(ctx context.Context, rw http.ResponseWriter, r *http.Request, _ httprouter.Params) (context.Context, int, error) {
		decoder := json.NewDecoder(r.Body)
		request, ok := reflect.New(reflect.TypeOf(requestType)).Interface().(RequestValidator)

		if !ok {
			return ctx, http.StatusInternalServerError, fmt.Errorf("Failed type assertion for type: %#v", requestType)
		}

		err := decoder.Decode(request)
		if err != nil {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return ctx, http.StatusInternalServerError, err
			}

			return ctx, http.StatusBadRequest, fmt.Errorf("Malformed request body: %s", string(body))
		}

		err = request.Validate()
		if err != nil {
			return ctx, http.StatusBadRequest, err
		}

		return context.WithValue(ctx, requestKey, request), 0, nil
	}
}
