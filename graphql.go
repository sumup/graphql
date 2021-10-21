// Package graphql provides a low level GraphQL client.
//
//  // create a client (safe to share across requests)
//  client := graphql.NewClient("https://machinebox.io/graphql")
//
//  // make a request
//  req := graphql.NewRequest(`
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type (
	// Client is a client for interacting with a GraphQL API.
	Client struct {
		endpoint         string
		httpClient       CustomHttpClient
		useMultipartForm bool

		// closeReq will close the request body immediately allowing for reuse of client
		closeReq bool

		// Log is called with various debug information.
		// To log to standard out, use:
		//  client.Log = func(s string) { log.Println(s) }
		Log func(s string)
	}

	// CustomHttpClient allows a custom http.Client to be used other than the default one provided by golang.
	CustomHttpClient interface {
		Do(*http.Request) (*http.Response, error)
	}

	// ClientOption are functions that are passed into NewClient to
	// modify the behaviour of the Client.
	ClientOption func(*Client)

	graphResponse struct {
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

// NewClient makes a new Client capable of making GraphQL requests.
// In case no option for http.Client is provided the default one is used in place.
func NewClient(endpoint string, opts ...ClientOption) *Client {
	c := &Client{
		endpoint: endpoint,
		Log:      func(string) {},
	}
	for _, optionFunc := range opts {
		optionFunc(c)
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c
}

// WithHTTPClient specifies the underlying http.Client to use when
// making requests.
//  NewClient(endpoint, WithHTTPClient(specificHTTPClient))
func WithHTTPClient(httpclient CustomHttpClient) ClientOption {
	return func(client *Client) {
		client.httpClient = httpclient
	}
}

// UseMultipartForm uses multipart/form-data and activates support for
// files.
func UseMultipartForm() ClientOption {
	return func(client *Client) {
		client.useMultipartForm = true
	}
}

//ImmediatelyCloseReqBody will close the req body immediately after each request body is ready
func ImmediatelyCloseReqBody() ClientOption {
	return func(client *Client) {
		client.closeReq = true
	}
}

func (c *Client) logf(format string, args ...interface{}) {
	c.Log(fmt.Sprintf(format, args...))
}

// Run executes the query and unmarshals the response from the data field
// into the response object.
// Pass in a nil response object to skip response parsing.
// If the request fails or the server returns an error, the first error
// will be returned.
func (c *Client) Run(ctx context.Context, op Operation, resp interface{}) Error {
	select {
	case <-ctx.Done():
		return NewExecutionError(ctx.Err())
	default:
	}
	if len(op.Files()) > 0 && !c.useMultipartForm {
		return NewExecutionError(errors.New("cannot send files with PostFields option"))
	}
	if c.useMultipartForm {
		return c.runWithPostFields(ctx, op, resp)
	}
	return c.runWithJSON(ctx, op, resp)
}

func (c *Client) runWithJSON(ctx context.Context, op Operation, resp interface{}) Error {
	req := op.Request()

	var requestBody bytes.Buffer
	requestBodyObj := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     req.q,
		Variables: req.vars,
	}
	if err := json.NewEncoder(&requestBody).Encode(requestBodyObj); err != nil {
		return NewExecutionError(errors.Wrap(err, "encode body"))
	}
	r, err := http.NewRequest(http.MethodPost, c.endpoint, &requestBody)
	if err != nil {
		return NewExecutionError(err)
	}
	r.Close = c.closeReq
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	r.Header.Set("Accept", "application/json; charset=utf-8")
	for key, values := range req.Header {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}
	c.logf(">> headers: %v", r.Header)
	r = r.WithContext(ctx)
	res, err := c.httpClient.Do(r)
	if err != nil {
		return NewExecutionError(err)
	}
	if res.StatusCode != http.StatusOK {
		return NewRequestError(res)
	}
	defer res.Body.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, res.Body); err != nil {
		return NewExecutionError(errors.Wrap(err, "reading body"))
	}
	c.logf("<< %s", buf.String())

	var gr *graphResponse
	switch op.(type) {
	case *Mutation:
		var results struct {
			Data map[string]graphMutationPayload
		}

		if err := json.NewDecoder(&buf).Decode(&results); err != nil {
			return NewExecutionError(errors.Wrap(err, "decoding response"))
		}
		gr = &graphResponse{}

		for _, result := range results.Data {
			if !result.Successful {
				messages := result.Messages
				errors := make([]GraphErr, len(messages))

				for i, message := range messages {
					errors[i] = GraphErr{
						Message: emptyOrString(message.Message),
						Code:    message.Code,
					}
				}

				gr.Errors = append(gr.Errors, errors...)
			} else {
				err := mapstructure.Decode(results.Data, &resp)
				if err != nil {
					return NewExecutionError(errors.Wrap(err, "decoding response"))
				}
			}
			// The code above only supports payloads with a single mutation
			break
		}

	default:
		gr = &graphResponse{Data: resp}
		if err := json.NewDecoder(&buf).Decode(&gr); err != nil {
			return NewExecutionError(errors.Wrap(err, "decoding response"))
		}
	}

	if len(gr.Errors) > 0 {
		return NewGraphQLError(gr.Errors, res)
	}
	return nil
}

func (c *Client) runWithPostFields(ctx context.Context, op Operation, resp interface{}) Error {
	req := op.Request()
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	if err := writer.WriteField("query", req.q); err != nil {
		return NewExecutionError(errors.Wrap(err, "write query field"))
	}
	var variablesBuf bytes.Buffer
	if len(req.vars) > 0 {
		variablesField, err := writer.CreateFormField("variables")
		if err != nil {
			return NewExecutionError(errors.Wrap(err, "create variables field"))
		}
		if err := json.NewEncoder(io.MultiWriter(variablesField, &variablesBuf)).Encode(req.vars); err != nil {
			return NewExecutionError(errors.Wrap(err, "encode variables"))
		}
	}
	for i := range req.files {
		part, err := writer.CreateFormFile(req.files[i].Field, req.files[i].Name)
		if err != nil {
			return NewExecutionError(errors.Wrap(err, "create form file"))
		}
		if _, err := io.Copy(part, req.files[i].R); err != nil {
			return NewExecutionError(errors.Wrap(err, "preparing file"))
		}
	}
	if err := writer.Close(); err != nil {
		return NewExecutionError(errors.Wrap(err, "close writer"))
	}
	c.logf(">> variables: %s", variablesBuf.String())
	c.logf(">> files: %d", len(req.files))
	c.logf(">> query: %s", req.q)
	gr := &graphResponse{
		Data: resp,
	}
	r, err := http.NewRequest(http.MethodPost, c.endpoint, &requestBody)
	if err != nil {
		return NewExecutionError(err)
	}
	r.Close = c.closeReq
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("Accept", "application/json; charset=utf-8")
	for key, values := range req.Header {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}
	c.logf(">> headers: %v", r.Header)
	r = r.WithContext(ctx)
	res, err := c.httpClient.Do(r)
	if err != nil {
		return NewExecutionError(err)
	}
	if res.StatusCode != http.StatusOK {
		return NewRequestError(res)
	}
	defer res.Body.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, res.Body); err != nil {
		return NewExecutionError(errors.Wrap(err, "reading body"))
	}
	c.logf("<< %s", buf.String())
	if err := json.NewDecoder(&buf).Decode(&gr); err != nil {
		return NewExecutionError(errors.Wrap(err, "decoding response"))
	}
	if len(gr.Errors) > 0 {
		return NewGraphQLError(gr.Errors, res)
	}
	return nil
}

func emptyOrString(pointer *string) string {
	if pointer == nil {
		return ""
	}
	return *pointer
}
