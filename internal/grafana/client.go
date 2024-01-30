package grafana

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/valyala/fasthttp"
)

// Client define an API client for grafana.
type Client struct {
	// Lock is required for request that require a specific user context.
	// https://grafana.com/docs/grafana/latest/developers/http_api/user/#switch-user-context-for-a-specified-user
	*sync.Mutex
	cfg config.Grafana
}

// Provide is a wire provider for Client.
func ProvideClient(cfg config.Grafana) Client {
	return Client{
		Mutex: &sync.Mutex{},
		cfg:   cfg,
	}
}

func (c Client) do(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error {
	client := fasthttp.Client{}
	deadline, hasDeadline := ctx.Deadline()

	var err error

	if hasDeadline {
		err = client.DoDeadline(req, resp, deadline)
	} else {
		err = client.Do(req, resp)
	}
	if err != nil {
		return err
	}

	// Follow redirect.
	for fasthttp.StatusCodeIsRedirect(resp.StatusCode()) {
		req.SetRequestURI(string(resp.Header.Peek("Location")))
		resp.Reset()

		if hasDeadline {
			err = client.DoDeadline(req, resp, deadline)
		} else {
			err = client.Do(req, resp)
		}
		if err != nil {
			return err
		}
	}

	return err
}

func (c Client) addAuthorizationHeader(req *fasthttp.Request) {
	basicStr := fmt.Sprintf("%s:%s", c.cfg.User.ExposeSecret(), c.cfg.Password.ExposeSecret())
	basicEncoded := base64.StdEncoding.EncodeToString([]byte(basicStr))

	req.Header.Set("Authorization", "Basic "+basicEncoded)
}
