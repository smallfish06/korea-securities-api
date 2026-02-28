package server

import (
	"net/http"

	"github.com/go-fuego/fuego"

	"github.com/smallfish06/krsec/pkg/broker"
)

// AuthTokenRequest represents an auth token request
type AuthTokenRequest struct {
	Broker      string             `json:"broker"`
	Credentials broker.Credentials `json:"credentials"`
	Sandbox     bool               `json:"sandbox"`
}

// handleAuthToken handles POST /auth/token
func (s *Server) handleAuthToken(c fuego.ContextWithBody[AuthTokenRequest]) (Response, error) {
	req, err := c.Body()
	if err != nil {
		return respond(c, http.StatusBadRequest, Response{
			OK:    false,
			Error: "invalid request body",
		})
	}

	brk := s.getFirstBroker()
	if brk == nil {
		return respond(c, http.StatusServiceUnavailable, Response{
			OK:    false,
			Error: "no broker available",
		})
	}

	token, err := brk.Authenticate(c.Context(), req.Credentials)
	if err != nil {
		return respond(c, statusFromBrokerError(err, http.StatusUnauthorized), Response{
			OK:    false,
			Error: err.Error(),
		})
	}

	return respond(c, http.StatusOK, Response{
		OK:     true,
		Data:   token,
		Broker: brk.Name(),
	})
}
