package kis

import (
	"time"
)

func newAuthedTestClient(baseURL string) *Client {
	c := NewClientWithTokenManager(false, nil)
	c.baseURL = baseURL
	c.SetCredentials("app", "secret")
	c.setToken("token", time.Now().Add(time.Hour))
	return c
}
