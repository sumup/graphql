package http

import (
	"io"
	"net/http"
	"regexp"
)

type addOperation struct {
	inner http.RoundTripper
}

const operationRegex = `((?:query\s)|(?:mutation\s))(\w+)`

func SetGraphqlOperation(inner http.RoundTripper) http.RoundTripper {
	return &addOperation{
		inner: inner,
	}
}

func (ug *addOperation) RoundTrip(r *http.Request) (*http.Response, error) {
	values := r.URL.Query()
	operation := getOperationName(r)

	if operation != "" {
		values.Add("operation", operation)
		r.URL.RawQuery = values.Encode()
	}

	return ug.inner.RoundTrip(r)
}

func getOperationName(r *http.Request) string {
	regex := regexp.MustCompile(operationRegex)
	getBody := r.GetBody

	copyBody, err := getBody()
	if err != nil {
		return ""
	}

	b, err := io.ReadAll(copyBody)
	if err != nil {
		return ""
	}

	operation := regex.FindAllSubmatch(b, -1)
	if operation != nil {
		return string(operation[0][2])
	}

	return ""
}
