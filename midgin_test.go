package midgin_test

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mwblythe/midgin"
	"github.com/stretchr/testify/suite"
)

var key1 = &struct{}{}

type MidginSuite struct {
	suite.Suite
}

func TestMidgin(t *testing.T) {
	suite.Run(t, &MidginSuite{})
}

func (s *MidginSuite) Router(middleware ...gin.HandlerFunc) *gin.Engine {
	router := gin.Default()

	router.Use(middleware...)

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}

func (s *MidginSuite) Request(router *gin.Engine) (http.ResponseWriter, *http.Request) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/ping", nil)
	s.Nil(err)
	router.ServeHTTP(w, r)

	s.Equal(http.StatusOK, w.Code)
	s.Equal("pong", w.Body.String())

	return w, r
}

func (s *MidginSuite) TestMiddleware1() {
	router := s.Router(midgin.Adapt(middleware1))
	_, r := s.Request(router)
	s.EqualValues("win", r.Context().Value(key1))
}

func (s *MidginSuite) TestMiddleware2() {
	router := s.Router(midgin.Adapt(middleware1), midgin.Adapt(middleware2))
	w, r := s.Request(router)

	s.EqualValues("win", r.Context().Value(key1))
	s.Equal("win", w.Header().Get("X-Midgen2"))
}

func (s *MidginSuite) TestMiddleware3() {
	router := s.Router(midgin.Adapt(middleware1), midgin.Adapt(middleware2), midgin.Adapt(middleware3))
	w, r := s.Request(router)

	s.EqualValues("win", r.Context().Value(key1))
	s.Equal("win", w.Header().Get("X-Midgen2"))
	s.Equal("win", w.Header().Get("X-Midgen3"))
}

// middleware1 sets a request context value
func middleware1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*r = *r.WithContext(context.WithValue(r.Context(), key1, "win"))
		next.ServeHTTP(w, r)
	})
}

// middleware2 wraps ResponseWriter in a basicWriter
func middleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&basicWriter{w}, r)
	})
}

// middleware3 wraps ResponseWriter in a fullWriter
func middleware3(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&fullWriter{w}, r)
	})
}

// basicWriter wraps a response writer
type basicWriter struct {
	http.ResponseWriter
}

func (w *basicWriter) WriteHeader(code int) {
	w.Header().Add("X-Midgen2", "win")
	w.ResponseWriter.WriteHeader(code)
}

// fullWriter wraps a response writer and implements all http interfaces
type fullWriter struct {
	http.ResponseWriter
}

func (w *fullWriter) WriteHeader(code int) {
	w.Header().Add("X-Midgen3", "win")
	w.ResponseWriter.WriteHeader(code)
}

func (w *fullWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *fullWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *fullWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify() // nolint
}
