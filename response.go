// Copyright 2020 dlboy(songdengtao). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nets

import (
	"bufio"
	"net"
	"net/http"
)

const (
	noWrittenSize = -1
	defaultStatus = http.StatusOK
)

// ResponseWriter ResponseWriter interface
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Returns the HTTP response status code of the current request.
	Status() int

	// Returns the number of bytes already written into the response http body.
	Size() int

	// Returns true if the response body was already written.
	Written() bool

	// get the http.Pusher for server push
	Pusher() http.Pusher

	// Forces to write the http header (status code + headers).
	WriteHeaderNow()
}

// responser http response writer
type responser struct {
	http.ResponseWriter
	size   int
	status int
}

// reset reset response
func (r *responser) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.status = defaultStatus
	r.size = noWrittenSize
}

// Status return the http response status
func (r *responser) Status() int {
	return r.status
}

// Size return the http response size
func (r *responser) Size() int {
	return r.size
}

// Written is already write http response header
func (r *responser) Written() bool {
	return r.size != noWrittenSize
}

// Write http.ResponseWriter.Write(data)
func (r *responser) Write(data []byte) (n int, err error) {
	n, err = r.ResponseWriter.Write(data)
	r.size += n
	return
}

// WriteHeader set status
func (r *responser) WriteHeader(code int) {
	if code > 0 && r.status != code {
		if r.Written() {
			debugPrintf("[WARNING] Headers were already written. Wanted to override status code %d with %d\n", r.status, code)
		}
		r.status = code
	}
}

// WriteHeaderNow Forces to write the http header (status code + headers).
func (r *responser) WriteHeaderNow() {
	if !r.Written() {
		r.size = 0
		r.ResponseWriter.WriteHeader(r.status)
	}
}

// Hijack implements the http.Hijacker interface.
func (r *responser) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if r.size < 0 {
		r.size = 0
	}
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotify interface.
func (r *responser) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flush interface.
func (r *responser) Flush() {
	r.WriteHeaderNow()
	r.ResponseWriter.(http.Flusher).Flush()
}

// Pusher return the http.PushPusher
func (r *responser) Pusher() (pusher http.Pusher) {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
