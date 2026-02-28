package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/smallfish06/krsec/internal/kiwoom"
	"github.com/smallfish06/krsec/pkg/broker"
)

type kiwoomProxyRequest struct {
	AccountID string                 `json:"account_id,omitempty"`
	Method    string                 `json:"method,omitempty"`
	APIID     string                 `json:"api_id"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Query     map[string]interface{} `json:"query,omitempty"`
	Body      map[string]interface{} `json:"body,omitempty"`
}

type kiwoomEndpointCaller interface {
	CallEndpoint(
		ctx context.Context,
		method string,
		path string,
		apiID string,
		fields map[string]string,
	) (map[string]interface{}, error)
}

// handleKiwoomProxy handles POST /kiwoom/{path...}
func (s *Server) handleKiwoomProxy(c fuego.ContextWithBody[kiwoomProxyRequest]) (Response, error) {
	rawPath := normalizeKiwoomProxyPath(c.PathParam("path"))
	if rawPath == "" {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "path is required"})
	}

	req, err := c.Body()
	if err != nil {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "invalid request body"})
	}

	apiID := strings.TrimSpace(req.APIID)
	if apiID == "" {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "api_id is required"})
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodPost
	}
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
	default:
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "unsupported method"})
	}

	brk, status, reason := s.resolveKiwoomProxyBroker(req.AccountID)
	if brk == nil {
		return respond(c, status, Response{OK: false, Error: reason})
	}

	impl, ok := brk.(kiwoomEndpointCaller)
	if !ok {
		return respond(c, http.StatusBadRequest, Response{OK: false, Error: "selected account does not support Kiwoom endpoint dispatch"})
	}

	fields := mergeStringMaps(
		mergeStringMaps(toStringMap(req.Query), toStringMap(req.Params)),
		toStringMap(req.Body),
	)
	result, err := impl.CallEndpoint(c.Context(), method, rawPath, apiID, fields)
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

func (s *Server) resolveKiwoomProxyBroker(accountID string) (broker.Broker, int, string) {
	accountID = strings.TrimSpace(accountID)
	if accountID != "" {
		brk, ok := s.getBrokerStrict(accountID)
		if !ok {
			return nil, http.StatusNotFound, "account not found"
		}
		if !strings.EqualFold(strings.TrimSpace(brk.Name()), broker.NameKiwoom) {
			return nil, http.StatusBadRequest, "account broker is not Kiwoom"
		}
		return brk, 0, ""
	}

	for _, acc := range s.accounts {
		if !strings.EqualFold(strings.TrimSpace(acc.Broker), broker.CodeKiwoom) {
			continue
		}
		if brk, ok := s.getBrokerStrict(acc.AccountID); ok {
			return brk, 0, ""
		}
	}
	return nil, http.StatusServiceUnavailable, "no Kiwoom account available"
}

func normalizeKiwoomProxyPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(path, kiwoom.PathPrefixAPISlash) {
		path = kiwoom.PathPrefixAPI + path
	}
	return path
}
