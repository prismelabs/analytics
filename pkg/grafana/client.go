package grafana

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/valyala/fasthttp"
)

const (
	GrafanaOrgIdHeader = "X-Grafana-Org-Id"
)

var (
	ErrGrafanaServerError = errors.New("grafana internal server error")
)

// Client define an API client for grafana.
type Client struct {
	cfg config.Grafana
}

// Provide is a wire provider for Client.
func ProvideClient(cfg config.Grafana) Client {
	return Client{
		cfg: cfg,
	}
}

func (c Client) do(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error {
	client := fasthttp.Client{
		DialDualStack: true,
	}
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

	if resp.StatusCode() >= 500 {
		return errors.Join(ErrGrafanaServerError, errors.New(string(resp.Body())))
	}

	return err
}

func (c Client) addAuthorizationHeader(req *fasthttp.Request) {
	basicStr := fmt.Sprintf("%s:%s", c.cfg.User.ExposeSecret(), c.cfg.Password.ExposeSecret())
	basicEncoded := base64.StdEncoding.EncodeToString([]byte(basicStr))

	req.Header.Set("Authorization", "Basic "+basicEncoded)
}
