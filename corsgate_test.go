package corsgate

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
})

func TestAllowSameOriginRequests(t *testing.T) {
	corsgate := New(Options{
		Origin:    []string{"localhost"},
		AllowSafe: false,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Origin", "localhost")

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestAllowSameOriginPOSTRequest(t *testing.T) {
	corsgate := New(Options{
		Origin:    []string{"localhost"},
		AllowSafe: false,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "localhost/", nil)
	req.Header.Add("Origin", "localhost")

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestAllowesPermittedCrossOriginRequests(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Origin", "localhost:1234")

	corsgate := New(Options{
		Origin:    []string{"localhost:1234"},
		AllowSafe: false,
	})

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestAllowesWildcardOrigins(t *testing.T) {
	corsgate := New(Options{
		Origin: []string{"*"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	req.Header.Add("Origin", "google.com")

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestRejectRequestsWithoutOrigin(t *testing.T) {
	corsgate := New(Options{
		Origin:    []string{"localhost"},
		AllowSafe: false,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusForbidden)
}

func TestRejectRequestsFromOtherOrigins(t *testing.T) {
	corsgate := New(Options{
		Origin:    []string{"localhost"},
		AllowSafe: false,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Origin", "google.com")

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusForbidden)
}

func TestAllowUnspecifiedSafeRequests(t *testing.T) {
	corsgate := New(Options{
		AllowSafe: true,
		Origin:    []string{"localhost"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestRejectRequestsWithoutOriginInStrictMode(t *testing.T) {
	corsgate := New(Options{
		AllowSafe: true,
		Strict:    true,
		Origin:    []string{"localhost"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusForbidden)
}

func TestFailureShouldNotBeInvokedForSameOriginRequests(t *testing.T) {
	corsgate := New(Options{
		AllowSafe: false,
		Origin:    []string{"localhost"},
		Failure: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Origin", "localhost")

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestFailureShouldBeInvokedForRequestsWithoutOrigin(t *testing.T) {
	corsgate := New(Options{
		AllowSafe: false,
		Origin:    []string{"localhost"},
		Failure: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusBadRequest)
}

func TestFallbackToRefererHeader(t *testing.T) {
	corsgate := New(Options{
		Origin:    []string{"localhost"},
		AllowSafe: false,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Referer", "http://localhost/mypage.html")

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestRefererShouldNotOverrideHost(t *testing.T) {
	corsgate := New(Options{
		Origin:    []string{"localhost"},
		AllowSafe: false,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Origin", "google.com")
	req.Header.Add("Referer", "http://localhost/mypage.html")

	corsgate.Handler(testHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusForbidden)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected [%v] (type %v) - Got [%v] (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
