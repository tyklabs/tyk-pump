package httpretry

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
)

type BackoffHTTPRetry struct {
	errMsg     string
	maxRetries uint64
	logger     *logrus.Entry
	httpclient *http.Client
}

type (
	conError     interface{ ConnectionError() bool }
	tempError    interface{ Temporary() bool }
	timeoutError interface{ Timeout() bool }
)

// NewBackoffRetry Creates an exponential backoff retry to use httpClient for connections. Will retry if a temporary error or
// 5xx or 429 status code in response.
func NewBackoffRetry(errMsg string, maxRetries uint64, httpClient *http.Client, logger *logrus.Entry) *BackoffHTTPRetry {
	return &BackoffHTTPRetry{errMsg: errMsg, maxRetries: maxRetries, httpclient: httpClient, logger: logger}
}

func (s *BackoffHTTPRetry) Send(req *http.Request) error {
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body.Close() // closing the original body
	}

	opFn := func() error {
		// recreating the request body from the buffer for each retry as if first attempt fails and
		// a new conn is created (keep alive disabled on server for example) the req body has already been read,
		// resulting in "http: ContentLength=X with Body length Y" error
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		resp, err := s.httpclient.Do(req)
		if err != nil {
			return s.handleErr(err)
		}
		defer func() {
			// read all response and discard so http client can
			// reuse connection as per doc on Response.Body
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("got status code %d and response '%s'", resp.StatusCode, body)

		// server error or rate limit hit - attempt retry
		if resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusTooManyRequests {
			return err
		}

		// any other error treat as permanent (i.e. auth error, invalid request) and don't retry
		return backoff.Permanent(err)
	}

	return backoff.RetryNotify(opFn, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), s.maxRetries), func(err error, t time.Duration) {
		s.logger.WithError(err).Errorf("%s retrying in %s", s.errMsg, t)
	})
}

func (s *BackoffHTTPRetry) handleErr(err error) error {
	if isErrorRetryable(err) {
		return err
	}
	// permanent error - don't retry
	return backoff.Permanent(err)
}

func isErrorRetryable(err error) bool {
	if err == nil {
		return false
	}

	var (
		conErr     conError
		tempErr    tempError
		timeoutErr timeoutError
		urlErr     *url.Error
		netOpErr   *net.OpError
	)

	switch {
	case errors.As(err, &conErr) && conErr.ConnectionError():
		return true

	case strings.Contains(err.Error(), "connection reset"):
		return true

	case errors.As(err, &urlErr):
		// Refused connections should be retried as the service may not yet be
		// running on the port. Go TCP dial considers refused connections as
		// not temporary.
		if strings.Contains(urlErr.Error(), "connection refused") {
			return true
		} else {
			return isErrorRetryable(errors.Unwrap(urlErr))
		}

	case errors.As(err, &netOpErr):
		// Network dial, or temporary network errors are always retryable.
		if strings.EqualFold(netOpErr.Op, "dial") || netOpErr.Temporary() {
			return true
		} else {
			return isErrorRetryable(errors.Unwrap(netOpErr))
		}

	case errors.As(err, &tempErr) && tempErr.Temporary():
		// Fallback to the generic temporary check, with temporary errors
		// retryable.
		return true

	case errors.As(err, &timeoutErr) && timeoutErr.Timeout():
		// Fallback to the generic timeout check, with timeout errors
		// retryable.
		return true
	}

	return false
}
