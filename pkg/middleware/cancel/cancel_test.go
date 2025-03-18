package cancel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kmmania/er_commonlib/pkg/middleware/cancel"
	"github.com/kmmania/er_commonlib/pkg/mocks/logger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func TestCancelUnaryInterceptor(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	tests := []struct {
		name        string
		ctx         context.Context
		handler     grpc.UnaryHandler
		expectedErr error
		expectedRet interface{}
	}{
		{
			name: "Context not canceled",
			ctx:  context.Background(),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			},
			expectedRet: "success",
		},
		/*		{
				name: "Context canceled",
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel() // Immediately cancel the context
					return ctx
				}(),
				handler: func(ctx context.Context, req interface{}) (interface{}, error) {
					return "should not reach here", nil
				},
				expectedErr: context.Canceled,
			},*/
		/*		{
				name: "Context canceled with deadline",
				ctx: func() context.Context {
					ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
					defer cancel()
					time.Sleep(10 * time.Millisecond) // Ensure deadline is exceeded
					return ctx
				}(),
				handler: func(ctx context.Context, req interface{}) (interface{}, error) {
					return "should not reach here", nil
				},
				expectedErr: context.DeadlineExceeded,
			},*/
		{
			name: "Handler returns error",
			ctx:  context.Background(),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, errors.New("handler error")
			},
			expectedErr: errors.New("handler error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := cancel.CancelUnaryInterceptor(env.mockLogger)
			ret, err := interceptor(tt.ctx, nil, nil, tt.handler)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedRet, ret)
		})
	}
}

func TestCancelStreamInterceptor(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	tests := []struct {
		name        string
		ctx         context.Context
		handler     grpc.StreamHandler
		expectedErr error
	}{
		{
			name: "Context not canceled",
			ctx:  context.Background(),
			handler: func(srv interface{}, ss grpc.ServerStream) error {
				return nil
			},
		},
		/*		{
				name: "Context canceled",
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				handler: func(srv interface{}, ss grpc.ServerStream) error {
					return nil // Should not be reached
				},
				expectedErr: context.Canceled,
			},*/
		/*		{
				name: "Context canceled with deadline",
				ctx: func() context.Context {
					ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
					defer cancel()
					time.Sleep(10 * time.Millisecond) // Ensure deadline is exceeded
					return ctx
				}(),
				handler: func(srv interface{}, ss grpc.ServerStream) error {
					return nil // Should not be reached
				},
				expectedErr: context.DeadlineExceeded,
			},*/
		{
			name: "Handler returns error",
			ctx:  context.Background(),
			handler: func(srv interface{}, ss grpc.ServerStream) error {
				return errors.New("handler error")
			},
			expectedErr: errors.New("handler error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := cancel.CancelStreamInterceptor(env.mockLogger)

			// Mock ServerStream to get the context
			mockServerStream := &mockServerStream{ctx: tt.ctx}

			err := interceptor(nil, mockServerStream, nil, tt.handler)

			assert.Equal(t, tt.expectedErr, err)
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
