package muxpromethus

import "net/http"

// statusResponseWriter captures the response status
type statusResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader satisifies the response writer interface
func (s *statusResponseWriter) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// Write satisifies the response writer interface
func (s *statusResponseWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = 200
	}
	n, err := s.ResponseWriter.Write(b)
	s.size += n
	return n, err
}
