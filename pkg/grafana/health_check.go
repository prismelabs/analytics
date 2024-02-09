package grafana

import (
	"context"
	"fmt"
	"time"

	"github.com/prismelabs/prismeanalytics/pkg/log"
	"github.com/valyala/fasthttp"
)

// HealthCheck performs an health check request against a grafana instance.
func (c Client) HealthCheck(ctx context.Context) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(c.cfg.Url + "/api/health")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana for health check: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to check grafana health: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// WaitHealthy checks grafana instance health status using the given config.
// This function panics if `maxRetry` attempt fails or are not healthy.
func WaitHealthy(logger log.Logger, c Client, maxRetry int) {
	var err error
	for retry := 0; retry < maxRetry; retry++ {
		logger.Info().
			Int("retry", retry).
			Int("max_retry", maxRetry).
			Msg("trying to connect to grafana")

		time.Sleep(time.Duration(retry) * time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.HealthCheck(ctx)
		if err != nil {
			continue
		}

		logger.Info().Msg("grafana connection established, service is healthy")
		break
	}

	if err != nil {
		logger.Panic().Msgf("failed to wait for healthy grafana: %v", err.Error())
	}
}
