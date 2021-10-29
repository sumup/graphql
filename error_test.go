package graphql

import (
	"net/http"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/pkg/errors"
)

func TestNewRequestError(t *testing.T) {
	is := is.New(t)
	response := &http.Response{}
	expected := RequestError{response: response}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(*err, expected)
}

func TestRequestErrorResponse(t *testing.T) {
	is := is.New(t)
	response := &http.Response{
		Status: http.StatusText(http.StatusNotFound),
	}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(err.Response(), response)
}

func TestRequestErrorError(t *testing.T) {
	is := is.New(t)
	response := &http.Response{
		Status: http.StatusText(http.StatusNotFound),
	}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(err.Error(), "request failed with status: Not Found")
}

func TestRequestErrorErrors(t *testing.T) {
	is := is.New(t)
	response := &http.Response{
		Status: http.StatusText(http.StatusNotFound),
	}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{"request failed with status: Not Found"})
}

func TestRequestErrorDetails(t *testing.T) {
	is := is.New(t)
	response := &http.Response{
		Status:     http.StatusText(http.StatusNotFound),
		StatusCode: http.StatusNotFound,
	}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(len(err.Details()), 1)
	is.Equal(err.Details()[0], ErrorDetail{
		Code:    http.StatusText(http.StatusNotFound),
		Message: "request failed with status: Not Found",
	})
}

func TestNewExecutionError(t *testing.T) {
	is := is.New(t)
	message := errors.New("some error")
	expected := ExecutionError{message: message}

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(*err, expected)
}

func TestExecutionErrorResponse(t *testing.T) {
	is := is.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(err.Response(), nil)
}

func TestExecutionErrorError(t *testing.T) {
	is := is.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(err.Error(), message.Error())
}

func TestExecutionErrorErrors(t *testing.T) {
	is := is.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{message.Error()})
}

func TestExecutionErrorDetails(t *testing.T) {
	is := is.New(t)
	message := errors.New("some error")

	err := NewExecutionError(message)

	is.True(err != nil)
	is.Equal(len(err.Details()), 1)
	is.Equal(err.Details()[0], ErrorDetail{
		Code:    "",
		Message: message.Error(),
	})
}

func TestNewGraphQLError(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}
	expected := GraphQLError{
		errors:   graphqlErrors,
		response: response,
	}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(*err, expected)
}

func TestGraphQLErrorResponse(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Response(), response)
}

func TestGraphQLErrorError(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Error(), graphqlErrors[1].Error())
}

func TestGraphQLErrorErrorEmptyList(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Error(), "")
}

func TestGraphQLErrorErrors(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{
		graphqlErrors[0].Error(),
		graphqlErrors[1].Error(),
	})
}

func TestGraphQLErrorCode(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{
		{
			Message: "secondary message",
			Extentions: GraphExt{
				Code: "ANOTHER_ERROR_CODE",
			},
		},
		{
			Code:    "ERROR_CODE",
			Message: "miscellaneous message as to why the the request was bad",
			Path:    []string{"field", "path"},
		},
	}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{
		graphqlErrors[0].Error(),
		graphqlErrors[1].Error(),
	})
	is.Equal(err.Code(), strings.ToLower(graphqlErrors[1].Code))
}

func TestGraphQLMutationError(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{
		{
			Message: "secondary message",
			Extentions: GraphExt{
				Code: "ANOTHER_ERROR_CODE",
			},
		},
		{
			Code:    "ERROR_CODE",
			Message: "miscellaneous message as to why the the request failed",
			Path:    []string{"field", "path"},
		},
	}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{
		graphqlErrors[0].Error(),
		graphqlErrors[1].Error(),
	})
	is.Equal(err.Code(), strings.ToLower(graphqlErrors[1].Code))
}

func TestGraphQLErrorDetails(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []GraphErr{
		{
			Message: "secondary message",
			Extentions: GraphExt{
				Code: "ANOTHER_ERROR_CODE",
			},
		},
		{
			Code:    "ERROR_CODE",
			Message: "miscellaneous message as to why the the request failed",
			Path:    []string{"field", "path"},
		},
	}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(len(err.Details()), 2)
	is.Equal(err.Details()[0], ErrorDetail{
		Code:    "another_error_code",
		Message: "secondary message",
		Domain:  "",
	})
	is.Equal(err.Details()[1], ErrorDetail{
		Code:    "error_code",
		Message: "miscellaneous message as to why the the request failed",
		Domain:  "field.path",
	})
}
