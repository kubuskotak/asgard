// Package rest is port adapter via http/s protocol
// # This manifest was generated by ymir. DO NOT EDIT.
package rest

import (
	"context"
	"net/http"
)

// StatusCreated error http StatusCreated.
func StatusCreated(r *http.Request) {
	*r = *r.WithContext(context.WithValue(r.Context(), CtxStatusCode, http.StatusCreated))
}

// StatusAccepted error http StatusAccepted.
func StatusAccepted(r *http.Request) {
	*r = *r.WithContext(context.WithValue(r.Context(), CtxStatusCode, http.StatusAccepted))
}

// StatusMovedPermanently error http StatusMovedPermanently.
func StatusMovedPermanently(r *http.Request) {
	*r = *r.WithContext(context.WithValue(r.Context(), CtxStatusCode, http.StatusMovedPermanently))
}
