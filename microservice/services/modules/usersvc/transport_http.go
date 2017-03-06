package usersvc

import (
	"encoding/json"
	"errors"
	"go/microservice/services/utils"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
)

// MakeHTTPHandler returns a handler that makes a set of endpoints available
// on predefined paths.
func MakeHTTPHandler(ctx context.Context, endpoints Endpoints, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}
	m := http.NewServeMux()
	m.Handle("/usersvc/GetUserinfo", httptransport.NewServer(
		ctx,
		endpoints.GetUserinfoEndpoint,
		DecodeHTTPGetUserinfoRequest,
		EncodeHTTPGetUserinfoResponse,
		options...,
	))
	return m
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	msg := err.Error()

	switch err {
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorWrapper{Error: msg})
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

// DecodeHTTPGetUserinfoRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded GetUserinfo request from the HTTP request body. Primarily useful in a
// server.
func DecodeHTTPGetUserinfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req getUserinfoRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func EncodeHTTPGetUserinfoResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	resp, ok := response.(getUserinfoResponse)
	if !ok {
		return json.NewEncoder(w).Encode(getUserinfoResponseHttp{
			Err: "interface conversion failed",
		})
	}

	return json.NewEncoder(w).Encode(getUserinfoResponseHttp{
		V:   resp.V,
		Err: utils.Err2Str(resp.Err),
	})
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGenericResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
