/*
Copyright 2024 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instantsnapshotgroup

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/oauth2"
)

type (
	httpResponse struct {
		url        string
		response   string
		statusCode int
	}

	mockToken struct {
		token *oauth2.Token
		err   error
	}
)

func (m *mockToken) Token() (*oauth2.Token, error) {
	return m.token, m.err
}

func TestToken(t *testing.T) {
	tests := []struct {
		name        string
		tokenGetter defaultTokenGetter
		wantErr     error
	}{
		{
			name: "ErrorToken",
			tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
				return &mockToken{
					token: nil,
					err:   cmpopts.AnyError,
				}, cmpopts.AnyError
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "Success",
			tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
				return &mockToken{
					token: &oauth2.Token{
						AccessToken: "access-token",
					},
					err: nil,
				}, nil
			},
			wantErr: nil,
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := token(ctx, tc.tokenGetter)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("token(%v) returned diff (-want +got):\n%s", tc.tokenGetter, diff)
			}
		})
	}
}

func TestGetResponseWithURLVariations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/success":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"key": "success_value"}`)
		case "/test/error":
			hj, _ := w.(http.Hijacker)
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		case "/test/illegal_bytes":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, []byte{0xFE, 0x0F})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	testCases := []struct {
		name    string
		s       *ISGService
		method  string
		baseURL string
		wantErr error
	}{
		{
			name:    "RequestCreationFailure",
			method:  "INVALID",
			baseURL: fmt.Sprintf("%c", 0x7f),
			wantErr: cmpopts.AnyError,
		},
		{
			name: "TokenErr",
			s: &ISGService{
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: nil,
						err:   cmpopts.AnyError,
					}, cmpopts.AnyError
				},
			},
			method:  "GET",
			baseURL: ts.URL + "/test/error",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "RequestError",
			s: &ISGService{
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			method:  "GET",
			baseURL: ts.URL + "/test/error",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "Success",
			s: &ISGService{
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			baseURL: ts.URL + "/test/success",
			wantErr: nil,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.s.GetResponse(ctx, tc.method, tc.baseURL, nil)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("GetResponse(%v, %v) returned diff (-want +got):\n%s", tc.method, tc.baseURL, diff)
			}
		})
	}
}

func TestGetProcessStatus(t *testing.T) {
	getProcessStatusHTTPHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/error":
			hj, _ := w.(http.Hijacker)
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		case "/test/invalid_json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"key": "success_value"`))
		case "/test/no_status":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"key": "success_value"}`))
		case "/test/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "RUNNING"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(getProcessStatusHTTPHandlerFunc))
	defer ts.Close()

	tests := []struct {
		s          *ISGService
		name       string
		baseURL    string
		wantStatus string
		wantErr    error
	}{
		{
			name: "InvalidRequest",
			s: &ISGService{
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			baseURL:    ts.URL + "/test/error",
			wantStatus: "",
			wantErr:    cmpopts.AnyError,
		},
		{
			name: "InvalidJSON",
			s: &ISGService{
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			baseURL:    ts.URL + "/test/invalid_json",
			wantStatus: "",
			wantErr:    cmpopts.AnyError,
		},
		{
			name: "NoStatus",
			s: &ISGService{
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			baseURL:    ts.URL + "/test/no_status",
			wantStatus: "",
			wantErr:    cmpopts.AnyError,
		},
		{
			name: "Success",
			s: &ISGService{
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			baseURL:    ts.URL + "/test/success",
			wantStatus: "RUNNING",
			wantErr:    nil,
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, err := tc.s.getProcessStatus(ctx, tc.baseURL)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("GetProcessStatus(%v) returned diff (-want +got):\n%s", tc.baseURL, diff)
			}
			if gotStatus != tc.wantStatus {
				t.Errorf("GetProcessStatus(%v) = %v, want: %v", tc.baseURL, gotStatus, tc.wantStatus)
			}
		})
	}
}

func TestCreateISGErrors(t *testing.T) {
	createISGHTTPHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/error":
			hj, _ := w.(http.Hijacker)
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		case "/test/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"key": "success_value"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(createISGHTTPHandlerFunc))
	defer ts.Close()

	tests := []struct {
		name    string
		s       *ISGService
		project string
		zone    string
		data    []byte
		wantErr error
	}{
		{
			name: "RequestError",
			s: &ISGService{
				baseURL:    ts.URL + "/test/error",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			data:    []byte(`{"sample_key": "sample_value"}`),
			wantErr: cmpopts.AnyError,
		},
		{
			name: "Success",
			s: &ISGService{
				baseURL:    ts.URL + "/test/success",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			data:    []byte(`{"sample_key": "sample_value"}`),
			wantErr: nil,
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.s.CreateISG(ctx, tc.project, tc.zone, tc.data)
			if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("CreateISG(%v, %v, %v) returned diff (-want +got):\n%s", tc.project, tc.zone, tc.data, diff)
			}
		})
	}
}

func TestIsgExistsErrors(t *testing.T) {
	isgExistsHTTPHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/error":
			hj, _ := w.(http.Hijacker)
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		case "/test/invalid_json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "INVALID"`))
		case "/test/deleting_status":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "DELETING"}`))
		case "/test/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "RUNNING"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(isgExistsHTTPHandlerFunc))
	defer ts.Close()

	tests := []struct {
		name    string
		s       *ISGService
		project string
		zone    string
		opName  string
		wantErr error
	}{
		{
			name: "RequestError",
			s: &ISGService{
				baseURL:    ts.URL + "/test/error",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			opName:  "test-op",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "InvalidJSON",
			s: &ISGService{
				baseURL:    ts.URL + "/test/invalid_json",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			opName:  "test-op",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "DeletingStatus",
			s: &ISGService{
				baseURL:    ts.URL + "/test/deleting_status",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			opName:  "test-op",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "Success",
			s: &ISGService{
				baseURL:    ts.URL + "/test/success",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			opName:  "test-op",
			wantErr: nil,
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.s.isgExists(ctx, tc.project, tc.zone, tc.opName)
			if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("IsgExists(%v, %v, %v) returned diff (-want +got):\n%s", tc.project, tc.zone, tc.opName, diff)
			}
		})
	}
}

func TestParseInstantSnapshotGroupURL(t *testing.T) {
	tests := []struct {
		name     string
		cgURL    string
		wantZone string
		wantCG   string
		wantErr  error
	}{
		{
			name:     "InvalidURL",
			cgURL:    "https://www.googleapis.com/compute/v1/projects/test-project/regions/test-region",
			wantZone: "",
			wantCG:   "",
			wantErr:  cmpopts.AnyError,
		},
		{
			name:     "Success",
			cgURL:    "https://www.googleapis.com/compute/v1/projects/test-project/zones/test-zone/instantSnapshotGroups/test-isg",
			wantZone: "test-zone",
			wantCG:   "test-isg",
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotZone, gotCG, err := parseInstantSnapshotGroupURL(tc.cgURL)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("parseInstantSnapshotGroupURL(%q) returned diff (-want +got):\n%s", tc.cgURL, diff)
			}
			if gotZone != tc.wantZone {
				t.Errorf("parseInstantSnapshotGroupURL(%q) = %q, want: %q", tc.cgURL, gotZone, tc.wantZone)
			}
			if gotCG != tc.wantCG {
				t.Errorf("parseInstantSnapshotGroupURL(%q) = %q, want: %q", tc.cgURL, gotCG, tc.wantCG)
			}
		})
	}
}

func TestDescribeInstantSnapshots(t *testing.T) {
	describeInstantSnapshotsHTTPHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/error":
			hj, _ := w.(http.Hijacker)
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		case "/test/invalid_json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "INVALID"`))
		case "/test/parse_isg_failure":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items": {"sourceInstantSnapshotGroup": "https://www.googleapis.com/compute/v1/projects/test-project/regions/test-region"}}`))
		case "/test/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items": [{"sourceInstantSnapshotGroup": "https://www.googleapis.com/compute/v1/projects/test-project/zones/test-zone/instantSnapshotGroups/test-isg"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(describeInstantSnapshotsHTTPHandlerFunc))
	defer ts.Close()

	tests := []struct {
		name    string
		s       *ISGService
		project string
		zone    string
		isg     string
		wantErr error
	}{
		{
			name: "RequestError",
			s: &ISGService{
				baseURL:    ts.URL + "/test/error",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			isg:     "test-isg",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "InvalidJSON",
			s: &ISGService{
				baseURL:    ts.URL + "/test/invalid_json",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			isg:     "test-isg",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "ParseISGFailure",
			s: &ISGService{
				baseURL:    ts.URL + "/test/parse_isg_failure",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			isg:     "test-isg",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "Success",
			s: &ISGService{
				baseURL:    ts.URL + "/test/success",
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			isg:     "test-isg",
			wantErr: nil,
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.s.DescribeInstantSnapshots(ctx, tc.project, tc.zone, tc.isg)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("DescribeInstantSnapshots(%v, %v, %v) returned diff (-want +got):\n%s", tc.project, tc.zone, tc.isg, diff)
			}
		})
	}
}

func TestDeleteISGErrors(t *testing.T) {
	deleteISGHTTPHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/error":
			hj, _ := w.(http.Hijacker)
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		case "/test/invalid_json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "INVALID"`))
		case "/test/delete_error":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ERROR"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(deleteISGHTTPHandlerFunc))
	defer ts.Close()

	tests := []struct {
		name    string
		s       *ISGService
		project string
		zone    string
		isgName string
		wantErr error
	}{
		{
			name: "RequestError",
			s: &ISGService{
				baseURL:    ts.URL + "/test/error",
				maxRetries: 1,
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			isgName: "test-isg",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "InvalidJSON",
			s: &ISGService{
				baseURL:    ts.URL + "/test/invalid_json",
				maxRetries: 1,
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			project: "test-project",
			zone:    "test-zone",
			isgName: "test-isg",
			wantErr: cmpopts.AnyError,
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.s.DeleteISG(ctx, tc.project, tc.zone, tc.isgName)
			if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("DeleteISG(%v, %v, %v) returned diff (-want +got):\n%s", tc.project, tc.zone, tc.isgName, diff)
			}
		})
	}
}

func TestWaitForISGUploadCompletionWithRetryErrors(t *testing.T) {
	isgUploadCompletionHTTPHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/error":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ERROR"}`))
		case "/test/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "READY"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(isgUploadCompletionHTTPHandlerFunc))
	defer ts.Close()

	tests := []struct {
		name    string
		s       *ISGService
		baseURL string
		wantErr error
	}{
		{
			name: "Error",
			s: &ISGService{
				maxRetries: 1,
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			baseURL: ts.URL + "/test/error",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "Success",
			s: &ISGService{
				maxRetries: 1,
				httpClient: defaultNewClient(10*time.Minute, defaultTransport()),
				tokenGetter: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
					return &mockToken{
						token: &oauth2.Token{
							AccessToken: "access-token",
						},
						err: nil,
					}, nil
				},
			},
			baseURL: ts.URL + "/test/success",
			wantErr: nil,
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.s.WaitForISGUploadCompletionWithRetry(ctx, tc.baseURL)
			if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("WaitForISGUploadCompletionWithRetry(%v) returned diff (-want +got):\n%s", tc.baseURL, diff)
			}
		})
	}
}
