package graphql

import (
	"fmt"
	"net/http"
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

	RequestError struct {
		response *http.Response
	}

	ExecutionError struct {
		message error
	}

	GraphQLError struct {
		errors   []GraphErr
		response *http.Response
	}

	GraphErr struct {
		Code       string
		Extentions GraphExt
		Message    string
		Path       []string
	}

	GraphExt struct {
		Code string
	}

	ErrorDetail struct {
		Code    string
		Message string
		Domain  string
	}
)

var (
	// Type assertions
	_ Error = &RequestError{}
	_ Error = &ExecutionError{}
	_ Error = &GraphQLError{}
)

func (e GraphErr) Error() string {
	message := e.Message
	if len(e.Path) > 0 {
		return e.ErrPath() + ": " + message
	}

	return message
}

func (e GraphErr) ErrCode() string {
	code := e.Code
	if len(code) > 0 {
		return strings.ToLower(code)
	}

	code = e.Extentions.Code
	if len(code) > 0 {
		return strings.ToLower(code)
	}

	return ""
}

func (e GraphErr) ErrPath() string {
	return strings.Join(e.Path, ".")
}

func (e GraphErr) ToErrorDetail() ErrorDetail {
	return ErrorDetail{
		Code:    e.ErrCode(),
		Message: e.Message,
		Domain:  e.ErrPath(),
	}
}

func NewRequestError(response *http.Response) *RequestError {
	return &RequestError{
		response: response,
	}
}

func (r *RequestError) Response() *http.Response {
	return r.response
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("request failed with status: %s", r.response.Status)
}

func (r *RequestError) Errors() []string {
	return []string{r.Error()}
}

func (r *RequestError) Code() string {
	return http.StatusText(r.response.StatusCode)
}

func (r *RequestError) Details() []ErrorDetail {
	return []ErrorDetail{
		{Code: r.Code(), Message: r.Error()},
	}
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

func (e *ExecutionError) Code() string {
	return ""
}

func (e *ExecutionError) Details() []ErrorDetail {
	return []ErrorDetail{
		{Code: e.Code(), Message: e.Error()},
	}
}

func NewGraphQLError(errors []GraphErr, response *http.Response) *GraphQLError {
	return &GraphQLError{
		errors:   errors,
		response: response,
	}
}

func (g *GraphQLError) Response() *http.Response {
	return g.response
}

func (g *GraphQLError) Code() string {
	errors := g.errors

	if len(errors) > 0 {
		return errors[len(errors)-1].ErrCode()
	}

	return ""
}

func (g *GraphQLError) Error() string {
	errors := g.errors

	if len(errors) > 0 {
		return errors[len(errors)-1].Error()
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

func (g *GraphQLError) Details() []ErrorDetail {
	errors := []ErrorDetail{}
	for _, err := range g.errors {
		errors = append(errors, err.ToErrorDetail())
	}

	return errors
}
