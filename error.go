package graphql

import (
	"net/http"
)

type (
	Error interface {
		error
		Response() *http.Response
		Errors() []string
	}

	RequestError struct {
		response *http.Response
	}

	ExecutionError struct {
		message error
	}

	GraphQLError struct {
		errors   []graphErr
		response *http.Response
	}
)

const (
	FailedHTTPRequestErrorMessage = "server returned a non-200 status code"
)

var (

	// Type assertions
	_ Error = &RequestError{}
	_ Error = &ExecutionError{}
	_ Error = &GraphQLError{}
)

func NewRequestError(response *http.Response) *RequestError {
	return &RequestError{
		response: response,
	}
}

func (r *RequestError) Response() *http.Response {
	return r.response
}

func (r *RequestError) Error() string {
	return FailedHTTPRequestErrorMessage
}

func (r *RequestError) Errors() []string {
	return []string{FailedHTTPRequestErrorMessage}
}

func NewExecutionError(message error) *ExecutionError {
	return &ExecutionError{
		message: message,
	}
}

func (e *ExecutionError) Response() *http.Response {
	return nil
}

func (e *ExecutionError) Error() string {
	return e.message.Error()
}

func (e *ExecutionError) Errors() []string {
	return []string{e.message.Error()}
}

func NewGraphQLError(errors []graphErr, response *http.Response) *GraphQLError {
	return &GraphQLError{
		errors:   errors,
		response: response,
	}
}

func (g *GraphQLError) Response() *http.Response {
	return g.response
}

func (g *GraphQLError) Error() string {
	if len(g.errors) > 0 {
		return g.errors[0].Error()
	}

	return ""
}

func (g *GraphQLError) Errors() []string {
	errors := []string{}
	for _, err := range g.errors {
		errors = append(errors, err.Error())
	}

	return errors
}
