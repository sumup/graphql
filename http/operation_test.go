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
	t.Run("SetGraphqlOperation success", func(t *testing.T) {
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.RawQuery, "operation=FooBar")

			_, _ = io.WriteString(w, "{}")
		})
		srv := httptest.NewServer(handlerFunc)
		defer srv.Close()

		transport := http.DefaultTransport
		transport = SetGraphqlOperation(transport)
		httpClient := &http.Client{Transport: transport}

		client := graphql.NewClient(srv.URL, graphql.WithHTTPClient(httpClient))

		ctx := context.Background()
		req := graphql.NewRequest("query FooBar {}")
		var resp struct {
			Value string
		}

		err := client.Run(ctx, req, &resp)
		assert.NoError(t, err)
	})
}
