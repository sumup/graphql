// Package graphql provides a low level GraphQL client.
//
//  // create a client (safe to share across requests)
//  client := graphql.NewClient("https://machinebox.io/graphql")
//
//  // make a defaultRequest
//  req := graphql.NewQueryOperation(`
//      query ($key: String!) {
//          items (id:$key) {
//              field1
//              field2
//              field3
//          }
//      }
//  `)
//
//  // set any variables
//  req.Var("key", "value")
//
//  // run it and capture the response
//  var respData ResponseStruct
//  if err := client.Run(ctx, req, &respData); err != nil {
//      log.Fatal(err)
//  }
//
// Specify client
//
// To specify your own http.Client, use the WithHTTPClient option:
//  httpclient := &http.Client{}
//  client := graphql.NewClient("https://machinebox.io/graphql", graphql.WithHTTPClient(httpclient))
package graphql

import (
	"context"
)

type (
	// Client is a client for interacting with a GraphQL API.
	Client struct {
		Chain RequestChain
	}

	GraphResponse struct {
		Data   interface{} `json:"data"`
		Errors []GraphErr  `json:"errors"`
	}

	graphValidationMessage struct {
		Code    string  `json:"code"`
		Message *string `json:"message"`
	}
	graphMutationPayload struct {
		Messages   []*graphValidationMessage `json:"messages"`
		Result     interface{}               `json:"result"`
		Successful bool                      `json:"successful"`
	}
)

// Deprecated: in favor of NewGraphClient
func NewClient(endpoint string, opts ...ChainOption) *Client {
	return NewGraphClient(NewChain(endpoint, opts...))
}

// NewGraphClient makes a new Client capable of making GraphQL requests.
// In case no option for http.Client is provided the default one is used in place.
func NewGraphClient(chain RequestChain) *Client {
	return &Client{
		Chain: chain,
	}
}

// Deprecated: Run executes the query and unmarshals the response from the data field
// into the response object.
// Pass in a nil response object to skip response parsing.
// If the defaultRequest fails or the server returns an error, the first error
// will be returned.
//
// See: Do
func (c *Client) Run(_ context.Context, op Operation, resp interface{}) Error {
	var i interface{} = op
	q, ok := i.(*queryOperation)
	if ok {
		q.ResponseType = resp
	}
	_, e := c.Do(op)
	return e
}

// Do executes the query and unmarshals the response from the data field
// into the expected response object.
// If the defaultRequest fails or the server returns an error, the first error
// will be returned.
func (c *Client) Do(op Operation) (*GraphResponse, Error) {
	return c.Chain.Action(op)
}
