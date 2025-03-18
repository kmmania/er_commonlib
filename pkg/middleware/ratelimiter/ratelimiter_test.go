package ratelimiter_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kmmania/er_commonlib/pkg/middleware/ratelimiter"
	"github.com/kmmania/er_commonlib/pkg/mocks/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type testEnv struct {
	ctrl       *gomock.Controller
	mockLogger *mocks.MockLogger
}

func setUpTestEnv(t *testing.T) *testEnv {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	return &testEnv{
		ctrl:       ctrl,
		mockLogger: mockLogger,
	}
}

func tearDownTestEnv(env *testEnv) {
	env.ctrl.Finish()
}

func TestRateLimiterHTTP(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	testCases := []struct {
		name      string
		rateLimit rate.Limit
		burst     int
		requests  []struct {
			expectStatus int
			expectBody   string
		}
	}{
		{
			name:      "Single request allowed, second request rate-limited",
			rateLimit: 1,
			burst:     1,
			requests: []struct {
				expectStatus int
				expectBody   string
			}{
				{expectStatus: http.StatusOK, expectBody: "OK"},
				{expectStatus: http.StatusTooManyRequests, expectBody: "Too many requests"},
			},
		},
		{
			name:      "Burst of 2 requests allowed, third request rate-limited",
			rateLimit: 1,
			burst:     2,
			requests: []struct {
				expectStatus int
				expectBody   string
			}{
				{expectStatus: http.StatusOK, expectBody: "OK"},
				{expectStatus: http.StatusOK, expectBody: "OK"},
				{expectStatus: http.StatusTooManyRequests, expectBody: "Too many requests"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl := rate.NewLimiter(tc.rateLimit, tc.burst)

			handler := func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			}

			router := gin.New()
			router.Use(ratelimiter.RateLimiterHTTP(rl, env.mockLogger))
			router.GET("/test", handler)

			for i, reqSpec := range tc.requests {
				t.Run(fmt.Sprintf("Request %d", i+1), func(t *testing.T) {
					req, _ := http.NewRequest("GET", "/test", nil)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)

					assert.Equal(t, reqSpec.expectStatus, w.Code)
					if reqSpec.expectBody != "" {
						assert.Contains(t, w.Body.String(), reqSpec.expectBody)
					}
				})
			}
		})
	}
}

func TestRateLimiterUnaryInterceptor(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	testCases := []struct {
		name      string
		rateLimit rate.Limit
		burst     int
		calls     []struct {
			expectError bool
			expectCode  codes.Code
			expectResp  string
		}
	}{
		{
			name:      "Single request allowed, second request rate-limited",
			rateLimit: 1,
			burst:     1,
			calls: []struct {
				expectError bool
				expectCode  codes.Code
				expectResp  string
			}{
				{expectError: false, expectCode: codes.OK, expectResp: "success"},
				{expectError: true, expectCode: codes.ResourceExhausted, expectResp: ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl := rate.NewLimiter(tc.rateLimit, tc.burst)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			}

			interceptor := ratelimiter.RateLimiterUnaryInterceptor(rl, env.mockLogger)

			for i, call := range tc.calls {
				t.Run(fmt.Sprintf("Call %d", i+1), func(t *testing.T) {
					resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)

					if call.expectError {
						assert.Error(t, err)
						st, ok := status.FromError(err)
						assert.True(t, ok)
						assert.Equal(t, call.expectCode, st.Code())
					} else {
						assert.NoError(t, err)
						assert.Equal(t, call.expectResp, resp)
					}
				})
			}
		})
	}
}

func TestRateLimiterStreamInterceptor(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	testCases := []struct {
		name      string
		rateLimit rate.Limit
		burst     int
		calls     []struct {
			expectError bool
			expectCode  codes.Code
		}
	}{
		{
			name:      "Single call allowed, second call rate-limited",
			rateLimit: 1,
			burst:     1,
			calls: []struct {
				expectError bool
				expectCode  codes.Code
			}{
				{expectError: false, expectCode: codes.OK},
				{expectError: true, expectCode: codes.ResourceExhausted},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl := rate.NewLimiter(tc.rateLimit, tc.burst)

			handler := func(srv interface{}, ss grpc.ServerStream) error {
				return nil
			}

			interceptor := ratelimiter.RateLimiterStreamInterceptor(rl, env.mockLogger)

			for i, call := range tc.calls {
				t.Run(fmt.Sprintf("Call %d", i+1), func(t *testing.T) {
					mockServerStream := &mockServerStream{ctx: context.Background()}

					err := interceptor(nil, mockServerStream, &grpc.StreamServerInfo{}, handler)

					if call.expectError {
						assert.Error(t, err)
						st, ok := status.FromError(err)
						assert.True(t, ok)
						assert.Equal(t, call.expectCode, st.Code())
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})
	}
}

type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) SendMsg(interface{}) error {
	return nil
}

func (m *mockServerStream) RecvMsg(interface{}) error {
	return nil
}

func (m *mockServerStream) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *mockServerStream) Trailer() metadata.MD {
	return nil
}

func (m *mockServerStream) CloseSend() error {
	return nil
}

func (m *mockServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetHeader(metadata.MD) error { // Added SetHeader
	return nil
}

func (m *mockServerStream) SetTrailer(metadata.MD) { // Added SetTrailer
}
