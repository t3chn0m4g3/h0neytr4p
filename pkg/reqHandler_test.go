package h0neytr4p

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func configureTestIO(t *testing.T) (string, string) {
	t.Helper()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "log.json")
	payloadDir := filepath.Join(dir, "payloads")
	if err := InitLogFile(logPath, false); err != nil {
		t.Fatalf("InitLogFile() error = %v", err)
	}
	t.Cleanup(func() {
		_ = CloseLogFile()
	})
	if err := InitPayloadFolder(payloadDir); err != nil {
		t.Fatalf("InitPayloadFolder() error = %v", err)
	}
	return logPath, payloadDir
}

func readTestLog(t *testing.T, logPath string) string {
	t.Helper()

	if err := CloseLogFile(); err != nil {
		t.Fatalf("CloseLogFile() error = %v", err)
	}
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", logPath, err)
	}
	return string(content)
}

func TestGetIPUsesFirstForwardedForAddress(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.50.20:54321"
	req.Header.Set("X-Forwarded-For", "10.0.0.5, 10.1.0.15, 10.2.0.14")

	if got, want := GetIP(req), "10.0.0.5"; got != want {
		t.Fatalf("GetIP() = %q, want %q", got, want)
	}
}

func TestGetIPFallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.50.20:54321"

	if got, want := GetIP(req), "192.168.50.20"; got != want {
		t.Fatalf("GetIP() = %q, want %q", got, want)
	}
}

func TestAllHandlerMatchesTextPlainParamsAfterPayloadCapture(t *testing.T) {
	logPath, _ := configureTestIO(t)
	trap := Trap{
		Basicinfo: BasicInfo{Name: "text-plain", Port: "80"},
		Behaviour: []Behaviour{
			{
				Request: Request{
					URL:     "/plain",
					Method:  http.MethodPost,
					Headers: map[string]interface{}{},
					Params:  map[string]interface{}{"token": "secret"},
				},
				Response: Response{
					Statuscode: http.StatusCreated,
					Body:       "matched",
					Headers:    map[string]interface{}{"Content-Type": "text/plain"},
					Type:       "string",
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/plain", strings.NewReader("token=secret"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	allHandler([]Trap{trap}, false).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("response status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if body := rec.Body.String(); body != "matched" {
		t.Fatalf("response body = %q, want %q", body, "matched")
	}
	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"trapped":"true"`) {
		t.Fatalf("log does not mark request as trapped: %s", logContent)
	}
	if !strings.Contains(logContent, `"payload":"token=secret"`) {
		t.Fatalf("log does not include text payload: %s", logContent)
	}
}

func TestAllHandlerMatchesJSONParamsAfterPayloadCapture(t *testing.T) {
	logPath, _ := configureTestIO(t)
	trap := Trap{
		Basicinfo: BasicInfo{Name: "json", Port: "80"},
		Behaviour: []Behaviour{
			{
				Request: Request{
					URL:     "/json",
					Method:  http.MethodPost,
					Headers: map[string]interface{}{"Content-Type": "application/json"},
					Params:  map[string]interface{}{"name": "spring.datasource.hikari.connection-init-sql", "value": "CREATE*"},
				},
				Response: Response{
					Statuscode: http.StatusOK,
					Body:       "json matched",
					Headers:    map[string]interface{}{},
					Type:       "string",
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(`{"name":"spring.datasource.hikari.connection-init-sql","value":"CREATE ALIAS"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	allHandler([]Trap{trap}, false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "json matched" {
		t.Fatalf("response body = %q, want %q", body, "json matched")
	}
	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"trapped":"true"`) {
		t.Fatalf("log does not mark JSON request as trapped: %s", logContent)
	}
	if !strings.Contains(logContent, `"payload":"{\"name\":\"spring.datasource.hikari.connection-init-sql\",\"value\":\"CREATE ALIAS\"}"`) {
		t.Fatalf("log does not include JSON payload: %s", logContent)
	}
}

func TestAllHandlerMatchesRequestProto(t *testing.T) {
	logPath, _ := configureTestIO(t)
	trap := Trap{
		Basicinfo: BasicInfo{Name: "http2-proto", Port: "443"},
		Behaviour: []Behaviour{
			{
				Request: Request{
					URL:     "/http2",
					Method:  http.MethodGet,
					Proto:   "HTTP/2*",
					Headers: map[string]interface{}{},
					Params:  map[string]interface{}{},
				},
				Response: Response{
					Statuscode: http.StatusOK,
					Body:       "http2 matched",
					Headers:    map[string]interface{}{},
					Type:       "string",
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/http2", nil)
	req.Proto = "HTTP/2.0"
	req.ProtoMajor = 2
	req.ProtoMinor = 0
	rec := httptest.NewRecorder()

	allHandler([]Trap{trap}, false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "http2 matched" {
		t.Fatalf("response body = %q, want %q", body, "http2 matched")
	}
	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"trapped":"true"`) {
		t.Fatalf("log does not mark HTTP/2 request as trapped: %s", logContent)
	}
	if !strings.Contains(logContent, `"request_proto":"HTTP/2.0"`) {
		t.Fatalf("log does not include request protocol: %s", logContent)
	}
}

func TestAllHandlerRejectsMismatchedRequestProto(t *testing.T) {
	logPath, _ := configureTestIO(t)
	trap := Trap{
		Basicinfo: BasicInfo{Name: "http2-proto", Port: "443"},
		Behaviour: []Behaviour{
			{
				Request: Request{
					URL:     "/http2",
					Method:  http.MethodGet,
					Proto:   "HTTP/2*",
					Headers: map[string]interface{}{},
					Params:  map[string]interface{}{},
				},
				Response: Response{
					Statuscode: http.StatusOK,
					Body:       "http2 matched",
					Headers:    map[string]interface{}{},
					Type:       "string",
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/http2", nil)
	rec := httptest.NewRecorder()

	allHandler([]Trap{trap}, false).ServeHTTP(rec, req)

	if body := rec.Body.String(); body != "" {
		t.Fatalf("response body = %q, want empty body for mismatched proto", body)
	}
	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"trapped":"false"`) {
		t.Fatalf("log does not mark HTTP/1.1 request as untrapped: %s", logContent)
	}
}

func TestAllHandlerMatchesDecodedBasicAuthorizationHeader(t *testing.T) {
	logPath, _ := configureTestIO(t)
	trap := Trap{
		Basicinfo: BasicInfo{Name: "decoded-basic", Port: "2087"},
		Behaviour: []Behaviour{
			{
				Request: Request{
					URL:    "/login/",
					Method: http.MethodPost,
					Headers: map[string]interface{}{
						"Authorization-Basic-Decoded": "*successful_internal_auth_with_timestamp*",
						"Cookie":                      "*whostmgrsession=*",
					},
					Params: map[string]interface{}{"login_only": "1"},
				},
				Response: Response{
					Statuscode: http.StatusOK,
					Body:       "decoded basic matched",
					Headers:    map[string]interface{}{},
					Type:       "string",
				},
			},
		},
	}

	decoded := "root:x\nsuccessful_internal_auth_with_timestamp=9999999999\nhasroot=1"
	req := httptest.NewRequest(http.MethodPost, "/login/?login_only=1", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(decoded)))
	req.Header.Set("Cookie", "whostmgrsession=example")
	rec := httptest.NewRecorder()

	allHandler([]Trap{trap}, false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "decoded basic matched" {
		t.Fatalf("response body = %q, want %q", body, "decoded basic matched")
	}
	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"trapped":"true"`) {
		t.Fatalf("log does not mark decoded Basic request as trapped: %s", logContent)
	}
	if strings.Contains(logContent, "successful_internal_auth_with_timestamp") {
		t.Fatalf("log leaks decoded Basic Authorization content: %s", logContent)
	}
}

func cve27654Trap() Trap {
	return Trap{
		Basicinfo: BasicInfo{
			Name:       "CVE-2026-27654",
			Port:       "443",
			RiskRating: "8.8",
		},
		Behaviour: []Behaviour{
			{
				Request: Request{
					URL:     "/*",
					Method:  "COPY",
					Headers: map[string]interface{}{"Destination": "*://*"},
					Params:  map[string]interface{}{},
				},
				Response: Response{
					Statuscode: http.StatusCreated,
					Body:       "",
					Headers:    map[string]interface{}{"Server": "nginx/1.28.2"},
					Type:       "string",
				},
			},
			{
				Request: Request{
					URL:     "/*",
					Method:  "MOVE",
					Headers: map[string]interface{}{"Destination": "*://*"},
					Params:  map[string]interface{}{},
				},
				Response: Response{
					Statuscode: http.StatusCreated,
					Body:       "",
					Headers:    map[string]interface{}{"Server": "nginx/1.28.2"},
					Type:       "string",
				},
			},
		},
	}
}

func TestAllHandlerMatchesCVE27654WebDAVMethods(t *testing.T) {
	for _, method := range []string{"COPY", "MOVE"} {
		t.Run(method, func(t *testing.T) {
			logPath, _ := configureTestIO(t)
			req := httptest.NewRequest(method, "/anything", nil)
			req.Header.Set("Destination", "https://example.test/dav/target")
			rec := httptest.NewRecorder()

			allHandler([]Trap{cve27654Trap()}, false).ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Fatalf("response status = %d, want %d", rec.Code, http.StatusCreated)
			}
			if server := rec.Header().Get("Server"); server != "nginx/1.28.2" {
				t.Fatalf("Server header = %q, want %q", server, "nginx/1.28.2")
			}
			logContent := readTestLog(t, logPath)
			if !strings.Contains(logContent, `"trapped":"true"`) {
				t.Fatalf("log does not mark WebDAV request as trapped: %s", logContent)
			}
			if !strings.Contains(logContent, `"trapped_for":"CVE-2026-27654"`) {
				t.Fatalf("log does not include CVE-2026-27654: %s", logContent)
			}
		})
	}
}

func TestAllHandlerRejectsCVE27654WithoutDestination(t *testing.T) {
	logPath, _ := configureTestIO(t)
	req := httptest.NewRequest("COPY", "/anything", nil)
	rec := httptest.NewRecorder()

	allHandler([]Trap{cve27654Trap()}, false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want default %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("response body = %q, want empty body for missing Destination", body)
	}
	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"trapped":"false"`) {
		t.Fatalf("log does not mark missing Destination request as untrapped: %s", logContent)
	}
	if strings.Contains(logContent, `"trapped_for":"CVE-2026-27654"`) {
		t.Fatalf("log includes CVE-2026-27654 for missing Destination: %s", logContent)
	}
}

func TestAllHandlerDoesNotCapturePayloadForDifferentMethodOnWildcardTrap(t *testing.T) {
	logPath, _ := configureTestIO(t)
	req := httptest.NewRequest(http.MethodPost, "/random", strings.NewReader("token=secret"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	allHandler([]Trap{cve27654Trap()}, false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want default %d", rec.Code, http.StatusOK)
	}
	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"trapped":"false"`) {
		t.Fatalf("log does not mark unmatched POST as untrapped: %s", logContent)
	}
	if strings.Contains(logContent, `"payload":"token=secret"`) {
		t.Fatalf("log includes payload captured through different-method wildcard trap: %s", logContent)
	}
}

func TestAllHandlerCapturesMultipartPayloadFile(t *testing.T) {
	logPath, payloadDir := configureTestIO(t)
	trap := Trap{
		Basicinfo: BasicInfo{Name: "multipart", Port: "80"},
		Behaviour: []Behaviour{
			{
				Request: Request{
					URL:     "/upload",
					Method:  http.MethodPost,
					Headers: map[string]interface{}{},
					Params:  map[string]interface{}{"id": "123"},
				},
				Response: Response{
					Statuscode: http.StatusOK,
					Body:       "ok",
					Headers:    map[string]interface{}{},
					Type:       "string",
				},
			},
		},
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("id", "123"); err != nil {
		t.Fatalf("WriteField() error = %v", err)
	}
	part, err := writer.CreateFormFile("file", "payload.txt")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
		t.Fatalf("multipart part Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart writer Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	allHandler([]Trap{trap}, false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", rec.Code, http.StatusOK)
	}
	payloadPath := filepath.Join(payloadDir, "5d41402abc4b2a76b9719d911017c592")
	info, err := os.Stat(payloadPath)
	if err != nil {
		t.Fatalf("saved payload stat error = %v", err)
	}
	if mode := info.Mode().Perm(); mode != RuntimeFileMode {
		t.Fatalf("saved payload mode = %o, want %o", mode, RuntimeFileMode)
	}

	logContent := readTestLog(t, logPath)
	if !strings.Contains(logContent, `"payload_hash_md5":"5d41402abc4b2a76b9719d911017c592"`) {
		t.Fatalf("log does not include payload hash: %s", logContent)
	}
	if !strings.Contains(logContent, `"payload_parameter":"id=123"`) {
		t.Fatalf("log does not include multipart field: %s", logContent)
	}
}
