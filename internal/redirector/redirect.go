package redirector

import (
	"errors"
	"fmt"
	"net/http"
)

type Redirect struct {
	Hostname string    `dynamodbav:"Hostname"`
	Type     Type      `dynamodbav:"Type"`
	Location string    `dynamodbav:"Location"`
	Status   Status    `dynamodbav:"Status"`
	Rewrites []Rewrite `dynamodbav:"Rewrites"`
}

func (r *Redirect) ServeHTTP(writer http.ResponseWriter, request *http.Request) error {
	if r.Location == "" {
		return errors.New("missing redirect location")
	}

	var location string
	path := request.URL.Path
	hasPathMatch := false
	if r.Rewrites != nil {
		for _, rewrite := range r.Rewrites {
			if rewrite.RegExp.Regexp == nil {
				continue
			}
			if rewrite.RegExp.MatchString(path) {
				hasPathMatch = true
				path = rewrite.RegExp.ReplaceAllString(path, rewrite.Replace)
				if rewrite.Final {
					break
				}
			}
		}
	}
	if hasPathMatch {
		location = fmt.Sprintf("https://%s%s", r.Location, path)
	} else {
		location = fmt.Sprintf("https://%s", r.Location)
	}

	if r.Type == TypeRedirect {
		writer.Header().Add("location", location)
		writer.WriteHeader(r.Status.StatusCode())
		return nil
	}
	return nil
}
