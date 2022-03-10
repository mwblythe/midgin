// Package midgin adapts standard net/http middleware for use in Gin
package midgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Adapt accepts a standard middleware and returns a Gin middleware
//
// r := gin.Default()
// r.Use(midgin.Adapt(middleware))
func Adapt(middleware func(next http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		stop := true

		middleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// reset the context's Writer in case the middleware wrapped it
				c.Writer = mkWriter(w, c.Writer)
				c.Request = r
				stop = false
				c.Next()
			}),
		).ServeHTTP(c.Writer, c.Request)

		if stop {
			c.Abort()
		}
	}
}

// writer implements the gin.ResponseWriter interface by splitting it into parts.
type writer struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier // nolint
	ginWriter
}

// ginWriter is the part of gin.ResponseWriter that is unique to Gin
type ginWriter interface {
	Status() int
	Size() int
	WriteString(string) (int, error)
	Written() bool
	WriteHeaderNow()
	Pusher() http.Pusher
}

// compose a gin.ResponseWriter
func mkWriter(rw http.ResponseWriter, gw gin.ResponseWriter) gin.ResponseWriter {
	if i, ok := rw.(gin.ResponseWriter); ok {
		return i
	}

	w := writer{
		ResponseWriter: rw,
		Hijacker:       gw,
		Flusher:        gw,
		CloseNotifier:  gw,
		ginWriter:      gw,
	}

	if i, ok := rw.(http.Hijacker); ok {
		w.Hijacker = i
	}
	if i, ok := rw.(http.Flusher); ok {
		w.Flusher = i
	}
	if i, ok := rw.(http.CloseNotifier); ok { //nolint
		w.CloseNotifier = i
	}

	return w
}
