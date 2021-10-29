package graphql

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestDoJSON(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		b, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":null}`+"\n")
		_, _ = io.WriteString(w, `{
			"data": {
				"something": "yes"
			}
		}`)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var responseData map[string]interface{}
	err := client.Run(ctx, NewRequest("query {}"), &responseData)
	is.NoErr(err)
	is.Equal(calls, 1) // calls
	is.Equal(responseData["something"], "yes")
}

func TestDoJSONServerError(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		b, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":null}`+"\n")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `Internal Server Error`)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var responseData map[string]interface{}
	err := client.Run(ctx, NewRequest("query {}"), &responseData)
	is.Equal(calls, 1) // calls
	is.Equal(err.Error(), "request failed with status: 500 Internal Server Error")
	is.Equal(err.Errors(), []string{"request failed with status: 500 Internal Server Error"})
	is.Equal(err.Response().StatusCode, http.StatusInternalServerError)
}

func TestDoJSONBadRequestErr(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		b, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":null}`+"\n")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
			"errors": [
				{
					"path": ["field", "path"],
					"message": "miscellaneous message as to why the the request was bad"
				}
			]
		}`)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var responseData map[string]interface{}
	query := NewRequest("query {}")
	err := client.Run(ctx, query, &responseData)
	is.Equal(calls, 1) // calls
	is.Equal(err.Error(), "miscellaneous message as to why the the request was bad")
	is.Equal(err.Errors(), []string{
		"miscellaneous message as to why the the request was bad",
	})
	is.Equal(err.Response().StatusCode, http.StatusOK)
	is.Equal(err.Details(), []ErrorDetail{
		{
			Code:    "",
			Message: "miscellaneous message as to why the the request was bad",
			Domain:  "field.path",
		},
	})
}

func TestQueryJSON(t *testing.T) {
	is := is.New(t)

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		b, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":{"username":"matryer"}}`+"\n")
		_, err = io.WriteString(w, `{"data":{"value":"some data"}}`)
		is.NoErr(err)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := NewClient(srv.URL)

	req := NewRequest("query {}")
	req.Var("username", "matryer")

	// check variables
	is.True(req != nil)
	is.Equal(req.Vars()["username"], "matryer") // nolint: staticcheck

	var resp struct {
		Value string
	}
	err := client.Run(ctx, req, &resp)
	is.NoErr(err)
	is.Equal(calls, 1)

	is.Equal(resp.Value, "some data")
}

func TestDoJSONMutation(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		_, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
	    "data": {
        "testMutation": {
          "messages": null,
          "result": {
						"name": "updated_name"
					},
          "successful": true
        }
	    }
		}`)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var responseData map[string]interface{}

	mutation := NewMutation(`
		mutation TestMudation($input:TestMutationInput!) {
			testMutation(input: $input) {
				successful
				result {
					name
				}
				messages {
					code
					message
				}
			}
		}
	`)
	err := client.Run(ctx, mutation, &responseData)
	is.Equal(calls, 1) // calls
	is.NoErr(err)
}

func TestDoJSONMutationWithStruct(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		_, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
	    "data": {
        "testMutation": {
          "messages": null,
          "result": {
						"name": "Darth Vader"
					},
          "successful": true
        }
	    }
		}`)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	type testMutationResult struct {
		Name string `json:"name"`
	}
	type testMutationPayload struct {
		Messages   []*graphValidationMessage `json:"messages"`
		Result     *testMutationResult       `json:"result"`
		Successful bool                      `json:"successful"`
	}

	var responseData struct {
		TestMutation testMutationPayload `json:"testMutation"`
	}

	mutation := NewMutation(`
		mutation TestMudation($input:TestMutationInput!) {
			testMutation(input: $input) {
				successful
				result {
					name
				}
				messages {
					code
					message
				}
			}
		}
	`)
	err := client.Run(ctx, mutation, &responseData)
	is.Equal(calls, 1) // calls
	is.NoErr(err)
	is.Equal("Darth Vader", responseData.TestMutation.Result.Name)
}

func TestDoJSONMutationErr(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		_, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
	    "data": {
        "testMutation": {
          "messages": [
            {
              "code": "internal_server_error",
              "message": "An error occurred"
            }
          ],
          "result": null,
          "successful": false
        }
	    }
		}`)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var responseData map[string]interface{}

	mutation := NewMutation(`
		mutation TestMudation($input:TestMutationInput!) {
			testMutation(input: $input) {
				successful
				result
				messages {
					code
					message
				}
			}
		}
	`)
	err := client.Run(ctx, mutation, &responseData)
	is.Equal(calls, 1) // calls
	is.Equal(err.Error(), "An error occurred")
	is.Equal(err.Errors(), []string{
		"An error occurred",
	})
	is.Equal(err.Response().StatusCode, http.StatusOK)
}

func TestHeader(t *testing.T) {
	is := is.New(t)

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Header.Get("X-Custom-Header"), "123")

		_, err := io.WriteString(w, `{"data":{"value":"some data"}}`)
		is.NoErr(err)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := NewClient(srv.URL)

	req := NewRequest("query {}")
	req.Header("X-Custom-Header", "123")

	var resp struct {
		Value string
	}
	err := client.Run(ctx, req, &resp)
	is.NoErr(err)
	is.Equal(calls, 1)

	is.Equal(resp.Value, "some data")
}
