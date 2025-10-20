package redirector

import (
	"errors"
	"net/http"
)

type Redirect struct {
	Hostname string `dynamodbav:"Hostname"`
	Type     Type   `dynamodbav:"Type"`
	Location string `dynamodbav:"Location"`
	Status   Status `dynamodbav:"Status"`
	// Rewrites []string `dynamodbav:"Rewrites,omitempty"`
}

func (r *Redirect) ServeHTTP(writer http.ResponseWriter, _ *http.Request) error {
	if r.Location == "" {
		return errors.New("missing redirect location")
	}
	if r.Type == TypeRedirect {
		writer.Header().Add("location", r.Location)
		writer.WriteHeader(r.Status.StatusCode())
		return nil
	}
	return nil
}
