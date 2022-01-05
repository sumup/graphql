package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type (

	// CustomHTTPClient allows a custom http.Client to be used other than the default one provided by golang.
	CustomHTTPClient interface {
		Do(*http.Request) (*http.Response, error)
	}

	// RequestChain is an interface provided by the graphql client
	// to give a view into the invocation chain. It's used to give
	// the ability of others to intercept the request and handle
	// the response for some very specific instrumentation.
	//
	// A RequestChain must be safe for concurrent use by multiple
	// goroutines.
	RequestChain interface {
		// Action executes a single HTTP transaction, returning
		// a Response for the provided Request.
		//
		// Action should not attempt to interpret the response. In
		// particular, Action must return err == nil if it obtained
		// a response, regardless of the response's HTTP status code.
		// A non-nil err should be reserved for failure to obtain a
		// response. Similarly, Action should not attempt to
		// handle higher-level protocol details such as redirects,
		// authentication, or cookies.
		//
		// Action should not modify the request, except for
		// consuming and closing the Request's Body. Action may
		// read fields of the request in a separate goroutine. Callers
		// should not mutate or reuse the request until the Response's
		// Body has been closed.
		//
		// Action must always close the body, including on errors,
		// but depending on the implementation may do so in a separate
		// goroutine even after Action returns. This means that
		// callers wanting to reuse the body for subsequent requests
		// must arrange to wait for the Close call before doing so.
		//
		// The Request's URL and Header fields must be initialized.
		Action(o Operation) (*GraphResponse, Error)
	}

	chain struct {
		// log is called with various debug information.
		// To log to standard out, use:
		//  client.log = func(s string) { log.Println(s) }
		log              func(s string)
		// closeReq will close the defaultRequest body immediately allowing for reuse of client
		closeReq         bool
		endpoint         string
		httpClient       CustomHTTPClient `default:"http.DefaultClient"`
		useMultipartForm bool
	}

	chainState struct {
		error     Error
		origin    *chain
		request   *http.Request
		response  *http.Response
		operation Operation
	}

	queryPayload struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	ChainOption func(*chain)
)

func NewChain(endpoint string, opts ...ChainOption) RequestChain {
	c := &chain{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
		log:        func(string) {},
	}
	for _, optionFunc := range opts {
		optionFunc(c)
	}
	return c
}

// WithHTTPClient specifies the underlying http.Client to use when
// making requests.
//  NewClient(endpoint, WithHTTPClient(specificHTTPClient))
func WithHTTPClient(httpclient CustomHTTPClient) ChainOption {
	return func(c *chain) {
		c.httpClient = httpclient
	}
}

// UseMultipartForm uses multipart/form-data and activates support for
// files.
func UseMultipartForm() ChainOption {
	return func(c *chain) {
		c.useMultipartForm = true
	}
}

// ImmediatelyCloseReqBody will close the req body immediately after each request body is ready
func ImmediatelyCloseReqBody() ChainOption {
	return func(c *chain) {
		c.closeReq = true
	}
}

func (d *chain) Action(o Operation) (*GraphResponse, Error) {
	return d.start(o).
		validate().
		createRequest().
		call().
		parseResponse()
}

func (d *chain) start(o Operation) *chainState {
	return &chainState{
		origin:    d,
		operation: o,
	}
}

func (c *chainState) validate() *chainState {
	if len(c.operation.Request().Files()) > 0 && !c.origin.useMultipartForm {
		c.error = NewExecutionError(errors.New("cannot send files with PostFields option"))
	}
	return c
}

func (c *chainState) createRequest() *chainState {
	if c.error == nil {
		request := c.operation.Request()
		requestBody, mimeType, err := c.createRequestBody()
		if err != nil {
			c.error = err
		}
		r, e := http.NewRequest(http.MethodPost, c.origin.endpoint, requestBody)
		if e != nil {
			c.error = NewExecutionError(e)
		}
		r.Close = c.origin.closeReq
		r.Header.Set("Content-Type", mimeType)
		r.Header.Set("Accept", "application/json; charset=utf-8")
		for key, values := range request.Headers() {
			for _, value := range values {
				r.Header.Add(key, value)
			}
		}
		c.logf(">> headers: %v", r.Header)
		c.request = r
	}
	return c
}

func (c *chainState) call() *chainState {
	if c.error == nil {
		response, err := c.origin.httpClient.Do(c.request)
		if err == nil {
			if response.StatusCode != http.StatusOK {
				c.error = NewRequestError(response)
			}
			c.response = response
		} else {
			c.error = NewExecutionError(err)
		}
	}
	return c
}

func (c *chainState) parseResponse() (*GraphResponse, Error) {
	if c.error == nil {
		defer c.response.Body.Close()
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, c.response.Body); err != nil {
			return nil, NewExecutionError(errors.Wrap(err, "reading body"))
		}
		c.logf("<< %s", buf.String())
		resp := c.operation.ResponseBodyAs()
		var gr *GraphResponse
		if c.operation.IsMutation() {
			var results struct {
				Data map[string]graphMutationPayload
			}

			if err := json.NewDecoder(&buf).Decode(&results); err != nil {
				return nil, NewExecutionError(errors.Wrap(err, "decoding response"))
			}
			gr = &GraphResponse{}

			for _, result := range results.Data {
				if !result.Successful {
					messages := result.Messages
					errs := make([]GraphErr, len(messages))

					for i, message := range messages {
						errs[i] = GraphErr{
							Message: emptyOrString(message.Message),
							Code:    message.Code,
						}
					}

					gr.Errors = append(gr.Errors, errs...)
				} else {
					err := mapstructure.Decode(results.Data, &resp)
					if err != nil {
						return nil, NewExecutionError(errors.Wrap(err, "decoding response"))
					}
				}
				// The code above only supports payloads with a single mutation
				break
			}
		} else {
			gr = &GraphResponse{Data: resp}
			if err := json.NewDecoder(&buf).Decode(&gr); err != nil {
				return nil, NewExecutionError(errors.Wrap(err, "decoding response"))
			}
		}
		if len(gr.Errors) > 0 {
			return nil, NewGraphQLError(gr.Errors, c.response)
		}
		return gr, nil
	}
	return nil, c.error
}

func (c *chainState) createRequestBody() (*bytes.Buffer, string, Error) {
	if c.origin.useMultipartForm {
		return c.createMultipartBody()
	}
	return c.createJSONBody()
}

func (c *chainState) createMultipartBody() (*bytes.Buffer, string, Error) {
	var requestBody bytes.Buffer
	request := c.operation.Request()
	writer := multipart.NewWriter(&requestBody)
	if err := writer.WriteField("query", request.Query()); err != nil {
		return nil, "", NewExecutionError(errors.Wrap(err, "write query field"))
	}
	var variablesBuf bytes.Buffer
	if len(request.Vars()) > 0 {
		variablesField, err := writer.CreateFormField("variables")
		if err != nil {
			return nil, "", NewExecutionError(errors.Wrap(err, "create variables field"))
		}
		if err := json.NewEncoder(io.MultiWriter(variablesField, &variablesBuf)).Encode(request.Vars()); err != nil {
			return nil, "", NewExecutionError(errors.Wrap(err, "encode variables"))
		}
	}
	files := request.Files()
	for i := range files {
		part, err := writer.CreateFormFile(files[i].Field, files[i].Name)
		if err != nil {
			return nil, "", NewExecutionError(errors.Wrap(err, "create form file"))
		}
		if _, err := io.Copy(part, files[i].R); err != nil {
			return nil, "", NewExecutionError(errors.Wrap(err, "preparing file"))
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", NewExecutionError(errors.Wrap(err, "close writer"))
	}
	c.logf(">> variables: %s", variablesBuf.String())
	c.logf(">> files: %c", len(files))
	c.logf(">> query: %s", request.Query())
	return &requestBody, writer.FormDataContentType(), nil
}

func (c *chainState) createJSONBody() (*bytes.Buffer, string, Error) {
	var requestBody bytes.Buffer
	request := c.operation.Request()
	requestBodyObj := &queryPayload{
		Query:     request.Query(),
		Variables: request.Vars(),
	}
	if err := json.NewEncoder(&requestBody).Encode(requestBodyObj); err != nil {
		return nil, "", NewExecutionError(errors.Wrap(err, "encode body"))
	}
	return &requestBody, "application/json; charset=utf-8", nil
}

func (c *chainState) logf(format string, args ...interface{}) {
	c.origin.log(fmt.Sprintf(format, args...))
}

func emptyOrString(pointer *string) string {
	if pointer == nil {
		return ""
	}
	return *pointer
}
