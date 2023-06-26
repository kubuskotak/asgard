// Package rest is port adapter via http/s protocol
// # This manifest was generated by ymir. DO NOT EDIT.
package rest

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

// ResponseCSV - Custom type to hold value from [][]string type to csv format response.
type ResponseCSV struct {
	Filename string
	Rows     [][]string
}

// ResponseType - Custom type to hold value for find and replace on context value response type.
type ResponseType int

// Declare related constants for each ResponseType starting with index 1.
const (
	CtxPagination ResponseType = iota
	CtxVersion
	CtxStatusCode
)

func (r ResponseType) String() string {
	return [...]string{
		"pagination-key",
		"version-key",
		"payload-key",
		"status-code-key",
	}[r]
}

// Index - Return index of the Constant.
func (r ResponseType) Index() int {
	return int(r)
}

// Meta holds the response definition for the Meta entity.
type Meta struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"error_message,omitempty"`
}

// Version holds the response definition for the Version entity.
type Version struct {
	Label  string `json:"label,omitempty"`
	Number string `json:"number,omitempty"`
}

// Pagination holds the response definition for the Pagination entity.
type Pagination struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"per_page,omitempty"`
	Size  int `json:"page_count,omitempty"`
	Total int `json:"total_count,omitempty"`
}

// Response holds the response definition for the Response entity.
type Response[W RequestConstraint, R ResponseConstraint] struct {
	Meta       `json:"meta"`
	Version    `json:"version"`
	Pagination `json:"pagination,omitempty"`
	Data       any `json:"data,omitempty"`
	next       Adapter[W, R]
}

func (e *Response[W, R]) processRequest(r *http.Request) error {
	var (
		zero        W
		binder, err = Bind(r, &zero)
	)
	if err != nil {
		return err
	}
	if err = binder.Validate(); err != nil {
		return err
	}
	*r = *r.WithContext(context.WithValue(r.Context(), CtxPayloadRequest, zero))
	return nil
}

func (e *Response[W, R]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errFunc := func(err error) {
		code, ok := r.Context().Value(CtxStatusCode).(int)
		if !ok || code < 1 {
			code = http.StatusInternalServerError
			*r = *r.WithContext(context.WithValue(r.Context(), CtxStatusCode, code))
		}
		w.Header().Set(HeaderContentType.String(), MIMEApplicationJSON.String())
		w.Header().Set(HeaderContentTypeOptions.String(), "nosniff")
		e.Meta = Meta{
			Code:    strconv.Itoa(code),
			Message: err.Error(),
		}
		b, err := json.Marshal(e)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	if ver, ok := r.Context().Value(CtxVersion).(Version); ok {
		e.Version = ver
	}
	if err := e.processRequest(r); err != nil {
		errFunc(err)
		return
	}
	payload, err := e.next(w, r)
	if err != nil {
		errFunc(err)
		return
	}
	if pagination, ok := r.Context().Value(CtxPagination).(Pagination); ok {
		e.Pagination = pagination
	}
	e.Data = payload
}

// JSON sends a JSON response with status code.
func (e *Response[W, R]) JSON(w http.ResponseWriter, r *http.Request) {
	e.Data = make(map[string]any) // reset data struct
	e.ServeHTTP(w, r)
	code, ok := r.Context().Value(CtxStatusCode).(int)
	if !ok || code < 1 {
		code = http.StatusOK
	}
	if code >= http.StatusBadRequest {
		return
	}
	w.Header().Set(HeaderContentType.String(), MIMEApplicationJSON.String())
	e.Meta = Meta{
		Code: strconv.Itoa(code),
	}
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(e); err != nil {
		log.Error().Err(ErrInternalServerError(w, r, err)).Msg("JSON")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Error().Err(ErrInternalServerError(w, r, err)).Msg("JSON")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}

// CSV sends a CSV format response with status code.
func (e *Response[W, R]) CSV(w http.ResponseWriter, r *http.Request) {
	e.ServeHTTP(w, r)
	code, ok := r.Context().Value(CtxStatusCode).(int)
	if !ok || code < 1 {
		code = http.StatusOK
	}
	if code >= http.StatusBadRequest {
		return
	}
	e.Meta = Meta{
		Code: strconv.Itoa(code),
	}

	if data, ok := e.Data.(ResponseCSV); ok {
		buf := &bytes.Buffer{}
		xCsv := csv.NewWriter(buf)
		for _, row := range data.Rows {
			if err := xCsv.Write(row); err != nil {
				w.Header().Set(HeaderContentTypeOptions.String(), "nosniff")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		xCsv.Flush()
		if err := xCsv.Error(); err != nil {
			w.Header().Set(HeaderContentTypeOptions.String(), "nosniff")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set(HeaderContentDesc.String(), "File Transfer")
		w.Header().Set(HeaderContentDisposition.String(), fmt.Sprintf("attachment; filename=%s.csv", data.Filename))
		w.Header().Set(HeaderContentType.String(), MIMETextCSVCharsetUTF8.String())

		w.WriteHeader(code)

		if _, err := w.Write(buf.Bytes()); err != nil {
			w.Header().Set(HeaderContentTypeOptions.String(), "nosniff")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	http.Error(w, http.ErrNotSupported.Error(), http.StatusBadRequest)
}

// Paging send a Pagination data.
func Paging(r *http.Request, p Pagination) {
	if p.Limit > 0 {
		p.Size = int(math.Round(float64(p.Total) / float64(p.Limit)))
	}
	*r = *r.WithContext(context.WithValue(r.Context(), CtxPagination, p))
}

// RequestNotFound request data not found.
type RequestNotFound struct{}

// ResponseNotFound response data not found.
type ResponseNotFound struct{}

// NotFoundDefault sets a custom http.HandlerFunc for routing paths that could
// not be found. The default 404 handler is `http.NotFound`.
func NotFoundDefault() http.HandlerFunc {
	return HandlerAdapter[RequestNotFound](func(w http.ResponseWriter, r *http.Request) (ResponseNotFound, error) {
		return ResponseNotFound{}, ErrNotFound(w, r, errors.New("resource is not found"))
	}).JSON
}
