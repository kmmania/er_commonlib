package timeout_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kmmania/er_commonlib/pkg/middleware/timeout"
	"github.com/kmmania/er_commonlib/pkg/mocks/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

func TestTimeoutMiddleware(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	testCases := []struct {
		name           string
		timeout        time.Duration
		sleepDuration  time.Duration
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Request within timeout",
			timeout:        100 * time.Millisecond,
			sleepDuration:  50 * time.Millisecond,
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Request exceeding timeout",
			timeout:        50 * time.Millisecond,
			sleepDuration:  100 * time.Millisecond,
			expectedStatus: http.StatusGatewayTimeout,
			expectedBody:   "Gateway Timeout{\"error\":\"request timeout\"}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(timeout.TimeoutMiddleware(tc.timeout, env.mockLogger))
			router.GET("/test", func(c *gin.Context) {
				select {
				case <-time.After(tc.sleepDuration):
					c.String(http.StatusOK, "OK")
				case <-c.Request.Context().Done():
					err := c.Request.Context().Err()
					if errors.Is(err, context.DeadlineExceeded) {
						c.String(http.StatusGatewayTimeout, "Gateway Timeout")
					} else {
						c.String(http.StatusInternalServerError, "Internal Server Error")
					}
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestTimeoutUnaryServerInterceptor(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	testCases := []struct {
		name          string
		timeout       time.Duration
		handlerDelay  time.Duration
		expectedResp  interface{}
		expectedError error
		expectedCode  codes.Code
	}{
		{
			name:          "Request within timeout",
			timeout:       100 * time.Millisecond,
			handlerDelay:  50 * time.Millisecond,
			expectedResp:  "OK",
			expectedError: nil,
			expectedCode:  codes.OK,
		},
		{
			name:          "Request exceeding timeout",
			timeout:       20 * time.Millisecond,
			handlerDelay:  100 * time.Millisecond,
			expectedResp:  nil,
			expectedError: status.Error(codes.DeadlineExceeded, "Deadline exceeded"),
			expectedCode:  codes.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				select {
				case <-time.After(tc.handlerDelay):
					return "OK", nil
				case <-ctx.Done():
					err := ctx.Err()
					if errors.Is(err, context.DeadlineExceeded) {
						return nil, status.Error(codes.DeadlineExceeded, "Deadline exceeded")
					}
					return nil, err
				}
			}

			interceptor := timeout.TimeoutUnaryServerInterceptor(tc.timeout, env.mockLogger)
			resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)

			if tc.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Equal(t, tc.expectedCode, status.Code(err))
			}
		})
	}
}

func TestTimeoutStreamServerInterceptor(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	testCases := []struct {
		name          string
		timeout       time.Duration
		handlerDelay  time.Duration
		expectedError error
		expectedCode  codes.Code
	}{
		{
			name:          "Request within timeout",
			timeout:       100 * time.Millisecond,
			handlerDelay:  50 * time.Millisecond,
			expectedError: nil,
			expectedCode:  codes.OK,
		},
		{
			name:          "Request exceeding timeout",
			timeout:       20 * time.Millisecond,
			handlerDelay:  100 * time.Millisecond,
			expectedError: status.Error(codes.DeadlineExceeded, "Deadline exceeded"),
			expectedCode:  codes.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStream := &mockGRPCServerStream{}

			handler := func(srv interface{}, stream grpc.ServerStream) error {
				ctx := stream.Context()
				select {
				case <-time.After(tc.handlerDelay):
					return nil
				case <-ctx.Done():
					err := ctx.Err()
					if errors.Is(err, context.DeadlineExceeded) {
						return status.Error(codes.DeadlineExceeded, "Deadline exceeded")
					}
					return err
				}
			}

			interceptor := timeout.TimeoutStreamServerInterceptor(tc.timeout, env.mockLogger)
			err := interceptor(nil, mockStream, &grpc.StreamServerInfo{}, handler)

			if tc.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedCode, status.Code(err))
			}
		})
	}
}

type mockGRPCServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockGRPCServerStream) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}
