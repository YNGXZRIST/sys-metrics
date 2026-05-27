package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mustParseCIDR(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("ParseCIDR(%q): %v", cidr, err)
	}
	return ipNet
}

func TestTrustedSubnet_nilSubnet(t *testing.T) {
	nextCalled := false
	h := TrustedSubnet(nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(XRealIP, "203.0.113.1")

	h.ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTrustedSubnet_ipInSubnet(t *testing.T) {
	ipNet := mustParseCIDR(t, "10.0.0.0/8")

	nextCalled := false
	h := TrustedSubnet(ipNet)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(XRealIP, "10.1.2.3")

	h.ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTrustedSubnet_ipOutsideSubnet(t *testing.T) {
	ipNet := mustParseCIDR(t, "10.0.0.0/8")

	h := TrustedSubnet(ipNet)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not run")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(XRealIP, "203.0.113.10")

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestTrustedSubnet_missingHeader(t *testing.T) {
	ipNet := mustParseCIDR(t, "10.0.0.0/8")

	nextCalled := false
	h := TrustedSubnet(ipNet)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h.ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTrustedSubnet_invalidIP(t *testing.T) {
	ipNet := mustParseCIDR(t, "10.0.0.0/8")

	nextCalled := false
	h := TrustedSubnet(ipNet)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(XRealIP, "not-an-ip")

	h.ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetHeaderXRealIP(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(XRealIP, "10.0.0.1")

		ip, ok := getHeaderXRealIP(req)
		if !ok || ip != "10.0.0.1" {
			t.Fatalf("got %q, %v; want 10.0.0.1, true", ip, ok)
		}
	})

	t.Run("missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		ip, ok := getHeaderXRealIP(req)
		if ok || ip != "" {
			t.Fatalf("got %q, %v; want empty, false", ip, ok)
		}
	})
}
