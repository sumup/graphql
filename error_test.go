package graphql

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	assertIs "github.com/matryer/is"
	"github.com/pkg/errors"
)

func TestNewRequestError(t *testing.T) {
	is := assertIs.New(t)
	response := &http.Response{}
	expected := HttpRequestError{response: response}

	err := NewHTTPRequestError(response)

	is.Equal(*err, expected)
}

func TestRequestErrorResponse(t *testing.T) {
	is := assertIs.New(t)
	response := &http.Response{
		Status: http.StatusText(http.StatusNotFound),
	}

	err := NewHTTPRequestError(response)

	is.True(err != nil)
	is.Equal(err.Response(), response) //nolint:bodyclose
}

func TestRequestErrorError(t *testing.T) {
	is := assertIs.New(t)
	response := &http.Response{
		Status: http.StatusText(http.StatusNotFound),
	}

	err := NewHTTPRequestError(response)

	is.True(err != nil)
	is.Equal(err.Error(), "defaultRequest failed with status: Not Found")
}

func TestRequestErrorErrors(t *testing.T) {
	is := assertIs.New(t)
	response := &http.Response{
		Status: http.StatusText(http.StatusNotFound),
	}

	err := NewHTTPRequestError(response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{"defaultRequest failed with status: Not Found"})
}

func TestRequestErrorDetails(t *testing.T) {
	is := assertIs.New(t)
	response := &http.Response{
		Status:     http.StatusText(http.StatusNotFound),
		StatusCode: http.StatusNotFound,
	}

	err := NewHTTPRequestError(response)

	is.True(err != nil)
	is.Equal(len(err.Details()), 1)
	is.Equal(err.Details()[0], &errorDetail{
		code:    strconv.Itoa(http.StatusNotFound),
		message: "defaultRequest failed with status: Not Found",
	})
}

func TestNewExecutionError(t *testing.T) {
	is := assertIs.New(t)
	message := errors.New("some error")
	expected := ExecutionError{message: message}

	err := NewExecutionError(message)

	is.Equal(*err, expected)
}

func TestExecutionErrorResponse(t *testing.T) {
	is := assertIs.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(err.Response(), nil) //nolint:bodyclose
}

func TestExecutionErrorError(t *testing.T) {
	is := assertIs.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(err.Error(), message.Error())
}

func TestExecutionErrorErrors(t *testing.T) {
	is := assertIs.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{message.Error()})
}

func TestExecutionErrorDetails(t *testing.T) {
	is := assertIs.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(len(err.Details()), 1)
	is.Equal(err.Details()[0], &errorDetail{
		code:    "",
		message: message.Error(),
	})
}

func TestNewGraphQLError(t *testing.T) {
	is := assertIs.New(t)
	graphqlErrors := []GraphError{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}
	expected := GraphRequestError{
		errors:   graphqlErrors,
		response: response,
	}

	err := NewGraphRequestError(graphqlErrors, response)

	is.Equal(*err, expected)
}

func TestGraphQLErrorResponse(t *testing.T) {
	is := assertIs.New(t)
	graphqlErrors := []GraphError{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}

	err := NewGraphRequestError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Response(), response) //nolint:bodyclose
}

func TestGraphQLErrorError(t *testing.T) {
	is := assertIs.New(t)
	graphqlErrors := []GraphError{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}

	err := NewGraphRequestError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Error(), graphqlErrors[1].Error())
}

func TestGraphQLErrorErrorEmptyList(t *testing.T) {
	is := assertIs.New(t)
	var graphqlErrors []GraphError
	response := &http.Response{}

	err := NewGraphRequestError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Error(), "")
}

func TestGraphQLErrorErrors(t *testing.T) {
	is := assertIs.New(t)
	graphqlErrors := []GraphError{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}

	err := NewGraphRequestError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{
		graphqlErrors[0].Error(),
		graphqlErrors[1].Error(),
	})
}

func TestGraphQLErrorCode(t *testing.T) {
	is := assertIs.New(t)
	graphqlErrors := []GraphError{
		{
			Message: "secondary message",
			Extensions: GraphExt{
				Code: "ANOTHER_ERROR_CODE",
			},
		},
		{
			Code:    "ERROR_CODE",
			Message: "miscellaneous message as to why the the defaultRequest was bad",
			Path:    []string{"field", "path"},
		},
	}
	response := &http.Response{}

	err := NewGraphRequestError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{
		graphqlErrors[0].Error(),
		graphqlErrors[1].Error(),
	})
	is.Equal(err.Code(), strings.ToLower(graphqlErrors[1].Code))
}

func TestGraphQLMutationError(t *testing.T) {
	is := assertIs.New(t)
	graphqlErrors := []GraphError{
		{
			Message: "secondary message",
			Extensions: GraphExt{
				Code: "ANOTHER_ERROR_CODE",
			},
		},
		{
			Code:    "ERROR_CODE",
			Message: "miscellaneous message as to why the the defaultRequest failed",
			Path:    []string{"field", "path"},
		},
	}
	response := &http.Response{}

	err := NewGraphRequestError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{
		graphqlErrors[0].Error(),
		graphqlErrors[1].Error(),
	})
	is.Equal(err.Code(), strings.ToLower(graphqlErrors[1].Code))
}

func TestGraphQLErrorDetails(t *testing.T) {
	is := assertIs.New(t)
	graphqlErrors := []GraphError{
		{
			Message: "secondary message",
			Extensions: GraphExt{
				Code: "ANOTHER_ERROR_CODE",
			},
		},
		{
			Code:    "ERROR_CODE",
			Message: "miscellaneous message as to why the the defaultRequest failed",
			Path:    []string{"field", "path"},
		},
	}
	response := &http.Response{}

	err := NewGraphRequestError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(len(err.Details()), 2)
	is.Equal(err.Details()[0], &errorDetail{
		code:    "another_error_code",
		message: "secondary message",
		domain:  "",
	})
	is.Equal(err.Details()[1], &errorDetail{
		code:    "error_code",
		message: "miscellaneous message as to why the the defaultRequest failed",
		domain:  "field.path",
	})
}
