package graphql

import (
	"io"
	"net/http"
)

// queryOperation is a GraphQL defaultRequest.
type (
	// Operation is a graphQL operation
	Operation interface {
		Request() GraphRequest
		ResponseBodyAs() interface{}
		IsMutation() bool
		Name() string
	}
	GraphRequest interface {
		Query() string
		File(string, string, io.Reader)
		Files() []File
		Var(string, interface{})
		Vars() map[string]interface{}
		Header(string, string)
		Headers() http.Header
	}
	// queryOperation is the regular graphQL query operation
	queryOperation struct {
		Req             GraphRequest
		ResponseType 	interface{}
		queryName       string
		// isMutation indicates if query is used to modify data in the data store
		isMutation		bool
	}
	defaultRequest struct {
		q     string
		vars  map[string]interface{}
		files []File
		// header represent any defaultRequest headers that will be set
		// when the defaultRequest is made.
		header http.Header
	}
	File interface {
		Field() string
		Name() string
		Reader() io.Reader
	}
	// file represents a file to upload.
	file struct {
		field  string
		name   string
		reader io.Reader
	}
)
// Deprecated: in favor of NewGraphOperation
func NewRequest(q string) Operation {
	return NewQueryOperation(q, "graphql", nil)
}

// NewQueryOperation creates a new graphql query operation
// Pass in a nil response object to skip response parsing.
func NewQueryOperation(query, queryName string, responseType interface{}) Operation {
	return &queryOperation{
		Req: newReq(query),
		ResponseType: responseType,
		queryName: queryName,
	}
}

// Deprecated: in favor of NewMutationOperation
func NewMutation(q string) Operation {
	return NewMutationOperation(q, "graphql", nil)
}

// NewMutationOperation creates a new graphql mutation operation
// Pass in a nil response object to skip response parsing.
func NewMutationOperation(query, queryName string, responseType interface{}) Operation {
	return &queryOperation{
		Req: newReq(query),
		ResponseType: responseType,
		isMutation: true,
		queryName: queryName,
	}
}

func (r *queryOperation) Name() string {
	return r.queryName
}

func (r *queryOperation) Request() GraphRequest {
	return r.Req
}

func (r *queryOperation) ResponseBodyAs() interface{} {
	return r.ResponseType
}

func (r *queryOperation) IsMutation() bool {
	return r.isMutation
}

func (d *defaultRequest) Var(key string, value interface{}) {
	if d.vars == nil {
		d.vars = make(map[string]interface{})
	}
	d.vars[key] = value
}

func (d *defaultRequest) Vars() map[string]interface{} {
	return d.vars
}

func (d *defaultRequest) Header(key, value string) {
	d.header.Set(key, value)
}

func (d *defaultRequest) Headers() http.Header {
	return d.header
}

// File sets a file to upload.
// Files are only supported with a Client that was created with
// the UseMultipartForm option.
func (d *defaultRequest) File(fieldName, fileName string, reader io.Reader) {
	d.files = append(d.files, &file{
		field: fieldName,
		name:  fileName,
		reader:     reader,
	})
}

func (d *defaultRequest) Files() []File {
	return d.files
}

// newReq makes a new queryOperation with the specified string.
func newReq(q string) GraphRequest {
	return &defaultRequest{
		q:      q,
		header: make(map[string][]string),
	}
}

// Query gets the query string of this defaultRequest.
func (d *defaultRequest) Query() string {
	return d.q
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Field() string {
	return f.field
}

func (f *file) Reader() io.Reader {
	return f.reader
}
