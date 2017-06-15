package corsgate

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Options that specify behaviour
type Options struct {
	Origin      []string
	AllowSafe   bool
	Strict      bool
	Failure     func(w http.ResponseWriter, r *http.Request)
	Credentials bool
}

// CORSGate struct to hold options
type CORSGate struct {
	options *Options
}

// New returns a new instance of CORSGate
func New(opts Options) *CORSGate {
	return &CORSGate{
		options: &opts,
	}
}

// Handler implements http.HandlerFunc so this lib can be implemented witht the native net/http lib
func (c *CORSGate) Handler(h http.Handler) http.Handler {
	// At least one origin is mandatory
	if len(c.options.Origin) < 1 {
		fmt.Errorf("must specify the server's origin")
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := c.Validate(w, r)

		if err != nil {
			if c.options.Failure != nil {
				c.options.Failure(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Validate - Validates CORS preferences for an incoming request
func (c *CORSGate) Validate(w http.ResponseWriter, r *http.Request) error {
	origin := strings.ToLower(r.Header.Get("origin"))

	if origin == "" {
		// Fallback to Referer header
		ref := strings.ToLower(r.Header.Get("referer"))
		if ref != "" {
			url, parseErr := url.Parse(ref)
			if parseErr != nil {
				return fmt.Errorf("tried to fallback to referer header but could not parse it")
			}

			origin = url.Host
		} else {
			if c.options.Strict || (!c.options.AllowSafe || ((r.Method != http.MethodGet) && r.Method != http.MethodHead)) {
				return fmt.Errorf("CORS validation failed. Supplied origin isn't safe and strict mode is enabled.")
			}

			return nil
		}
	}

	// Always allow same-origin requests
	for _, host := range c.options.Origin {
		if host == origin {
			return nil
		} else if host == "*" {
			return nil
		}
	}

	// Check if we already have a access-control-allow-origin header
	curHeader := w.Header().Get("access-control-allow-origin")
	if curHeader != "" {
		curHeader = strings.ToLower(strings.Replace(curHeader, " ", "", -1))
		if (curHeader == "*") || (origin == curHeader) {
			return nil
		}
	}

	return fmt.Errorf("%s is not an allowed origin", origin)
}
