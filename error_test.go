package graphql

import (
	"net/http"
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
	response := &http.Response{}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(err.Response(), response)
}

func TestRequestErrorError(t *testing.T) {
	is := is.New(t)
	response := &http.Response{}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(err.Error(), FailedHTTPRequestErrorMessage)
}

func TestRequestErrorErrors(t *testing.T) {
	is := is.New(t)
	response := &http.Response{}

	err := NewRequestError(response)

	is.True(err != nil)
	is.Equal(err.Errors(), []string{FailedHTTPRequestErrorMessage})
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

func TestNewGraphQLError(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []graphErr{
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
	graphqlErrors := []graphErr{
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
	graphqlErrors := []graphErr{
		{Message: "some error"},
		{Message: "other error"},
	}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Error(), graphqlErrors[0].Error())
}

func TestGraphQLErrorErrorEmptyList(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []graphErr{}
	response := &http.Response{}

	err := NewGraphQLError(graphqlErrors, response)

	is.True(err != nil)
	is.Equal(err.Error(), "")
}

func TestGraphQLErrorErrors(t *testing.T) {
	is := is.New(t)
	graphqlErrors := []graphErr{
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
