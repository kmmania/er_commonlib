package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kmmania/er_backend/er_lib/pkg/cache"
	mocks "github.com/kmmania/er_backend/er_lib/pkg/mocks/cache"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testEnv struct {
	ctrl      *gomock.Controller
	mockCache *mocks.MockRedisCache
}

func setUpTestEnv(t *testing.T) *testEnv {
	ctrl := gomock.NewController(t)
	mockCache := mocks.NewMockRedisCache(ctrl)
	return &testEnv{
		ctrl:      ctrl,
		mockCache: mockCache,
	}
}

func tearDownTestEnv(env *testEnv) {
	env.ctrl.Finish()
}

func TestRedisCache_Get(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	type testCase struct {
		name        string
		setupMock   func()
		key         string
		timeout     time.Duration
		expectedErr error
		expectedVal map[string]interface{}
	}

	tests := []testCase{
		{
			name: "Success - Value Found",
			setupMock: func() {
				env.mockCache.EXPECT().
					Get(gomock.Any(), "key1", gomock.Any(), time.Second).
					DoAndReturn(func(ctx context.Context, key string, dest interface{}, timeout time.Duration) error {
						*(dest.(*map[string]interface{})) = map[string]interface{}{"value": 42}
						return nil
					})
			},
			key:         "key1",
			timeout:     time.Second,
			expectedErr: nil,
			expectedVal: map[string]interface{}{"value": 42},
		},
		{
			name: "Failure - Cache Miss",
			setupMock: func() {
				env.mockCache.EXPECT().
					Get(gomock.Any(), "key2", gomock.Any(), time.Second).
					Return(cache.ErrCacheMiss)
			},
			key:         "key2",
			timeout:     time.Second,
			expectedErr: cache.ErrCacheMiss,
			expectedVal: nil,
		},
		{
			name: "Failure - Redis Error",
			setupMock: func() {
				env.mockCache.EXPECT().
					Get(gomock.Any(), "key3", gomock.Any(), time.Second).
					Return(errors.New("redis error"))
			},
			key:         "key3",
			timeout:     time.Second,
			expectedErr: errors.New("redis error"),
			expectedVal: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			var result map[string]interface{}
			err := env.mockCache.Get(context.Background(), tc.key, &result, tc.timeout)
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.expectedVal, result)
			}
		})
	}
}

func TestRedisCache_Set(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	type testCase struct {
		name      string
		setupMock func()
		key       string
		value     map[string]interface{}
		ttl       time.Duration
		timeout   time.Duration
	}

	tests := []testCase{
		{
			name: "Success - Set Value",
			setupMock: func() {
				env.mockCache.EXPECT().
					Set(gomock.Any(), "key1", map[string]interface{}{"value": 42}, time.Minute, time.Second).
					Times(1) // Nous vérifions uniquement que l'appel est fait.
			},
			key:     "key1",
			value:   map[string]interface{}{"value": 42},
			ttl:     time.Minute,
			timeout: time.Second,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			env.mockCache.Set(context.Background(), tc.key, tc.value, tc.ttl, tc.timeout)
		})
	}
}

func TestRedisCache_Delete(t *testing.T) {
	env := setUpTestEnv(t)
	defer tearDownTestEnv(env)

	type testCase struct {
		name        string
		setupMock   func()
		key         string
		timeout     time.Duration
		expectedErr error
	}

	tests := []testCase{
		{
			name: "Success - Delete Key",
			setupMock: func() {
				env.mockCache.EXPECT().
					Delete(gomock.Any(), "key1", time.Second).
					Return(nil)
			},
			key:         "key1",
			timeout:     time.Second,
			expectedErr: nil,
		},
		{
			name: "Failure - Redis Error",
			setupMock: func() {
				env.mockCache.EXPECT().
					Delete(gomock.Any(), "key2", time.Second).
					Return(errors.New("redis error"))
			},
			key:         "key2",
			timeout:     time.Second,
			expectedErr: errors.New("redis error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			err := env.mockCache.Delete(context.Background(), tc.key, tc.timeout)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
