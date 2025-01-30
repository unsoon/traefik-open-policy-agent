package traefik_open_policy_agent

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dghubble/sling"
	"github.com/unsoon/traefik-open-policy-agent/helpers"
)

type Config struct {
	Url           string        `json:"url,omitempty"`
	AllowField    string        `json:"allowField,omitempty"`
	ErrorResponse ErrorResponse `json:"errorResponse,omitempty"`
}

type ErrorResponse struct {
	Headers     map[string]string `json:"headers,omitempty"`
	StatusCode  int               `json:"statusCode,omitempty"`
	ContentType string            `json:"contentType,omitempty"`
	Body        *interface{}      `json:"body,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		AllowField: "allow",
		ErrorResponse: ErrorResponse{
			StatusCode:  http.StatusUnauthorized,
			ContentType: "text/plain",
		},
	}
}

type OpenPolicyAgent struct {
	next          http.Handler
	url           string
	allowField    string
	errorResponse ErrorResponse
	name          string
}

type OpenPolicyAgentInput struct {
	Host    string              `json:"host"`
	Path    []string            `json:"path"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
	Body    json.RawMessage     `json:"body"`
}

type OpenPolicyAgentPayload struct {
	Input OpenPolicyAgentInput `json:"input"`
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &OpenPolicyAgent{
		next:          next,
		url:           config.Url,
		allowField:    config.AllowField,
		errorResponse: config.ErrorResponse,
		name:          name,
	}, nil
}

type OpenPolicyAgentResponse struct {
	Result map[string]interface{} `json:"result"`
}

func (h *OpenPolicyAgent) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	payload := requestToOpenPolicyAgentPayload(req)

	var response OpenPolicyAgentResponse

	if _, err := sling.New().Post(h.url).BodyJSON(payload).Receive(&response, "_"); err != nil {
		h.writeErrorResponse(rw)
		return
	}

	if isAllowed, isPresent := response.Result[h.allowField]; !isPresent || !isAllowed.(bool) {
		h.writeErrorResponse(rw)
		return
	}

	h.next.ServeHTTP(rw, req)
}

func (o *OpenPolicyAgent) writeErrorResponse(rw http.ResponseWriter) {
	for key, value := range o.errorResponse.Headers {
		rw.Header().Set(key, value)
	}

	rw.Header().Set("Content-Type", o.errorResponse.ContentType)
	rw.WriteHeader(o.errorResponse.StatusCode)

	if o.errorResponse.Body != nil {
		body, err := helpers.ConvertToType(*o.errorResponse.Body, o.errorResponse.ContentType)

		if err != nil {
			rw.Write([]byte(err.Error()))
			return
		}

		rw.Write([]byte(body))
	}
}

func requestToOpenPolicyAgentPayload(req *http.Request) OpenPolicyAgentPayload {
	var body json.RawMessage
	if req.Body != nil {
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		body = json.RawMessage(bodyBytes)
	}

	return OpenPolicyAgentPayload{
		Input: OpenPolicyAgentInput{
			Host:    req.Host,
			Path:    strings.Split(req.URL.Path, "/")[1:],
			Method:  req.Method,
			Headers: req.Header,
			Query:   req.URL.Query(),
			Body:    body,
		},
	}
}
