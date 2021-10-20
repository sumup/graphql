package graphql

import (
	"io"
	"net/http"
)

// Request is a GraphQL request.
type (
	Operation interface {
		Request() *Req
		File(string, string, io.Reader)
		Files() []File
		Var(string, interface{})
		Vars() map[string]interface{}
		Header(string, string)
		Headers() http.Header
	}
	Request struct {
		Req *Req
	}
	Mutation struct {
		Req *Req
	}
	Req struct {
		q     string
		vars  map[string]interface{}
		files []File
		// Header represent any request headers that will be set
		// when the request is made.
		Header http.Header
	}

	// File represents a file to upload.
	File struct {
		Field string
		Name  string
		R     io.Reader
	}
)

func NewRequest(q string) *Request {
	return &Request{
		Req: newReq(q),
	}
}

func NewMutation(m string) *Mutation {
	return &Mutation{
		Req: newReq(m),
	}
}

func (q *Request) Request() *Req {
	return q.Req
}

func (r *Request) Var(key string, value interface{}) {
	r.Request().Var(key, value)
}

func (r *Request) Vars() map[string]interface{} {
	return r.Request().Vars()
}

func (r *Request) Header(key, value string) {
	r.Request().Header.Set(key, value)
}

func (r *Request) Headers() http.Header {
	return r.Request().Header
}

func (r *Request) File(fieldname, filename string, reader io.Reader) {
	r.Req.File(fieldname, filename, reader)
}

func (r *Request) Files() []File {
	return r.Req.files
}

func (m *Mutation) Request() *Req {
	return m.Req
}

func (m *Mutation) Var(key string, value interface{}) {
	m.Request().Var(key, value)
}

func (m *Mutation) Vars() map[string]interface{} {
	return m.Request().Vars()
}

func (m *Mutation) Header(key, value string) {
	m.Request().Header.Set(key, value)
}

func (m *Mutation) Headers() http.Header {
	return m.Request().Header
}

func (m *Mutation) File(fieldname, filename string, reader io.Reader) {
	m.Req.File(fieldname, filename, reader)
}

func (m *Mutation) Files() []File {
	return m.Req.files
}

// NewRequest makes a new Request with the specified string.
func newReq(q string) *Req {
	req := &Req{
		q:      q,
		Header: make(map[string][]string),
	}
	return req
}

// Var sets a variable.
func (req *Req) Var(key string, value interface{}) {
	if req.vars == nil {
		req.vars = make(map[string]interface{})
	}
	req.vars[key] = value
}

// Vars gets the variables for this Request.
func (req *Req) Vars() map[string]interface{} {
	return req.vars
}

// Files gets the files in this request.
func (req *Req) Files() []File {
	return req.files
}

// Query gets the query string of this request.
func (req *Req) Query() string {
	return req.q
}

// File sets a file to upload.
// Files are only supported with a Client that was created with
// the UseMultipartForm option.
func (req *Req) File(fieldname, filename string, reader io.Reader) {
	req.files = append(req.files, File{
		Field: fieldname,
		Name:  filename,
		R:     reader,
	})
}
