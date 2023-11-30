package graphql

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type (
	Error interface {
		error
		Response() *http.Response
		Errors() []string
		Code() string
		Details() []ErrorDetail
	}

	// HttpRequestError occurs as an error response from a HTTP call.
	HttpRequestError struct {
		response *http.Response
	}

	// ExecutionError means one of 2 things: 1. request failed for some reason and wasn't
	// fulfilled; 2. request was fulfilled and a response was given but post response
	// processing failed.
	ExecutionError struct {
		message error
		response *http.Response
	}

	// GraphRequestError happens when errors were found at the graphql layer.
	GraphRequestError struct {
		errors   []GraphError
		response *http.Response
	}

	GraphError struct {
		Code       string
		Extensions GraphExt
		Message    string
		Path       []string
	}

	GraphExt struct {
		Code string
	}

	ErrorDetail interface {
		Code()    string
		Message() string
		Domain()  string
	}

	errorDetail struct {
		code    string
		message string
		domain  string
	}
)

var (
	// Type assertions
	_ Error = &HttpRequestError{}
	_ Error = &ExecutionError{}
	_ Error = &GraphRequestError{}
)

func (e GraphError) Error() string {
	return e.Message
}

func (e GraphError) ErrCode() string {
	code := e.Code
	if len(code) > 0 {
		return strings.ToLower(code)
	}

	code = e.Extensions.Code
	if len(code) > 0 {
		return strings.ToLower(code)
	}

	return ""
}

func (e GraphError) ErrPath() string {
	return strings.Join(e.Path, ".")
}


func (e *errorDetail) Code() string {
	return e.code
}

func (e *errorDetail) Message() string {
	return e.message
}

func (e *errorDetail) Domain()  string {
	return e.domain
}

func (e GraphError) ToErrorDetail() ErrorDetail {
	return &errorDetail{
		code:    e.ErrCode(),
		message: e.Message,
		domain:  e.ErrPath(),
	}
}

func NewHTTPRequestError(response *http.Response) *HttpRequestError {
	return &HttpRequestError{
		response: response,
	}
}

func (r *HttpRequestError) Response() *http.Response {
	return r.response
}

func (r *HttpRequestError) Error() string {
	return fmt.Sprintf("defaultRequest failed with status: %s", r.response.Status)
}

func (r *HttpRequestError) Errors() []string {
	return []string{r.Error()}
}

func (r *HttpRequestError) Code() string {
	if r.response != nil {
		return strconv.Itoa(r.response.StatusCode)
	}
	return ""
}

func (r *HttpRequestError) Details() []ErrorDetail {
	var e []ErrorDetail
	e = append(e, &errorDetail{
		code: r.Code(),
		message: r.Error(),
	})
	return e
}

func NewExecutionError(message error) *ExecutionError {
	return NewExecutionResponseError(message, nil)
}

func NewExecutionResponseError(message error, response *http.Response) *ExecutionError {
	return &ExecutionError{
		message: message,
		response: response,
	}
}

func (e *ExecutionError) Response() *http.Response {
	return e.response
}

func (e *ExecutionError) Error() string {
	return e.message.Error()
}

func (e *ExecutionError) Errors() []string {
	return []string{e.message.Error()}
}

func (e *ExecutionError) Code() string {
	if e.response != nil {
		return strconv.Itoa(e.response.StatusCode)
	}
	return ""
}

func (e *ExecutionError) Details() []ErrorDetail {
	var ed []ErrorDetail
	ed = append(ed, &errorDetail{
		code: e.Code(),
		message: e.Error(),
	})
	return ed
}

func NewGraphRequestError(errors []GraphError, response *http.Response) *GraphRequestError {
	return &GraphRequestError{
		errors:   errors,
		response: response,
	}
}

func (g *GraphRequestError) Response() *http.Response {
	return g.response
}

func (g *GraphRequestError) Code() string {
	errors := g.errors

	if len(errors) > 0 {
		return errors[len(errors)-1].ErrCode()
	}

	return ""
}

func (g *GraphRequestError) Error() string {
	errors := g.errors

	if len(errors) > 0 {
		return errors[len(errors)-1].Error()
	}

	return ""
}

func (g *GraphRequestError) Errors() []string {
	var errors []string
	for _, err := range g.errors {
		errors = append(errors, err.Error())
	}

	return errors
}

func (g *GraphRequestError) Details() []ErrorDetail {
	var errors []ErrorDetail
	for _, err := range g.errors {
		errors = append(errors, err.ToErrorDetail())
	}

	return errors
}
