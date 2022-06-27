package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	graphql "github.com/sumup/graphql"
)

func Test_SetGraphqlOperation(t *testing.T) {
	t.Run("SetGraphqlOperation with query", func(t *testing.T) {
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.RawQuery, "operation=FooBar")

			_, _ = io.WriteString(w, "{}")
		})

		runner := func(t *testing.T, client *graphql.Client) {
			ctx := context.Background()
			req := graphql.NewRequest("query FooBar {}")
			var resp struct{}

			err := client.Run(ctx, req, &resp)
			assert.NoError(t, err)
		}

		runTest(t, handlerFunc, runner)
	})

	t.Run("SetGraphqlOperation with mutation", func(t *testing.T) {
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.RawQuery, "operation=FooBar")

			_, _ = io.WriteString(w, `{
				"data": {
		        "fooBar": {
		            "messages": [],
		            "result": null,
		            "successful": false
		        }
		    }
			}`)
		})

		runner := func(t *testing.T, client *graphql.Client) {
			ctx := context.Background()
			req := graphql.NewMutation(`
				mutation FooBar {
					fooBar {
				    successful
				    messages
				    result
					}
				}
			`)
			var resp struct{}

			err := client.Run(ctx, req, &resp)
			assert.NoError(t, err)
		}

		runTest(t, handlerFunc, runner)
	})

func runTest(t *testing.T, handlerFunc http.HandlerFunc, runner func(*testing.T, *graphql.Client)) {
	srv := httptest.NewServer(handlerFunc)
	defer srv.Close()

	transport := http.DefaultTransport
	transport = SetGraphqlOperation(transport)
	httpClient := &http.Client{Transport: transport}

	client := graphql.NewClient(srv.URL, graphql.WithHTTPClient(httpClient))

	runner(t, client)
}
