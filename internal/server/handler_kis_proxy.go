package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"

	"github.com/smallfish06/krsec/internal/kis"
	"github.com/smallfish06/krsec/pkg/broker"
)

type kisProxyRequest struct {
	AccountID string                 `json:"account_id,omitempty"`
	Method    string                 `json:"method,omitempty"`
	TRID      string                 `json:"tr_id"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Query     map[string]interface{} `json:"query,omitempty"`
	Body      map[string]interface{} `json:"body,omitempty"`
}

type kisEndpointCaller interface {
	CallEndpoint(
		ctx context.Context,
		method string,
		path string,
		trID string,
		fields map[string]string,
	) (map[string]interface{}, error)
}

// handleKISProxy handles POST /kis/{path...}
func (s *Server) handleKISProxy(c fuego.ContextWithBody[kisProxyRequest]) (Response, error) {
	rawPath := normalizeKISProxyPath(c.PathParam("path"))
	if rawPath == "" {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "path is required"})
	}

	req, err := c.Body()
	if err != nil {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "invalid request body"})
	}

	trID := strings.TrimSpace(req.TRID)
	if trID == "" {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "tr_id is required"})
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
	default:
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "unsupported method"})
	}

	brk, status, reason := s.resolveKISProxyBroker(req.AccountID)
	if brk == nil {
		return respond(c, status, Response{OK: false, Error: reason})
	}

	impl, ok := brk.(kisEndpointCaller)
	if !ok {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "selected account does not support KIS endpoint dispatch"})
	}

	fields := mergeStringMaps(
		mergeStringMaps(toStringMap(req.Query), toStringMap(req.Params)),
		toStringMap(req.Body),
	)
	result, err := impl.CallEndpoint(c.Context(), method, rawPath, trID, fields)
	if err != nil {
		return respond(c, statusFromBrokerError(err, http.StatusInternalServerError), Response{
			OK:     false,
			Error:  err.Error(),
			Broker: brk.Name(),
		})
	}

	return respond(c, http.StatusOK, Response{
		OK:     true,
		Data:   result,
		Broker: brk.Name(),
	})
}

func (s *Server) resolveKISProxyBroker(accountID string) (broker.Broker, int, string) {
	accountID = strings.TrimSpace(accountID)
	if accountID != "" {
		brk, ok := s.getBrokerStrict(accountID)
		if !ok {
			return nil, http.StatusNotFound, "account not found"
		}
		if !strings.EqualFold(strings.TrimSpace(brk.Name()), broker.NameKIS) {
			return nil, http.StatusBadRequest, "account broker is not KIS"
		}
		return brk, 0, ""
	}

	for _, acc := range s.accounts {
		if !strings.EqualFold(strings.TrimSpace(acc.Broker), broker.CodeKIS) {
			continue
		}
		if brk, ok := s.getBrokerStrict(acc.AccountID); ok {
			return brk, 0, ""
		}
	}
	return nil, http.StatusServiceUnavailable, "no KIS account available"
}

func normalizeKISProxyPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(path, kis.PathPrefixUAPISlash) {
		path = kis.PathPrefixUAPI + path
	}
	return path
}

func toStringMap(src map[string]interface{}) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		if v == nil {
			out[key] = ""
			continue
		}
		out[key] = fmt.Sprint(v)
	}
	return out
}

func mergeStringMaps(base map[string]string, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}
