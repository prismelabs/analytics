package originregistry

import (
	"context"
	"io"
	"testing"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	logger := log.New("env_var_service_test", io.Discard, false)

	t.Run("NewService", func(t *testing.T) {
		t.Run("Error", func(t *testing.T) {
			t.Run(".fr", func(t *testing.T) {
				service, err := NewService(Config{Origins: []string{".com"}}, logger)
				require.Error(t, err)
				require.Nil(t, service)
			})
			t.Run("foo.fr", func(t *testing.T) {
				service, err := NewService(Config{Origins: []string{"foo..fr"}}, logger)
				require.Error(t, err)
				require.Nil(t, service)
			})
			t.Run("**.foo.fr", func(t *testing.T) {
				service, err := NewService(Config{Origins: []string{"**.foo.fr"}}, logger)
				require.Error(t, err)
				require.Nil(t, service)
			})
			t.Run("*foo.fr", func(t *testing.T) {
				service, err := NewService(Config{Origins: []string{"*foo.fr"}}, logger)
				require.Error(t, err)
				require.Nil(t, service)
			})
			t.Run("bar.*.foo.fr", func(t *testing.T) {
				service, err := NewService(Config{Origins: []string{"*foo.fr"}}, logger)
				require.Error(t, err)
				require.Nil(t, service)
			})
			t.Run("NoOrigin", func(t *testing.T) {
				service, err := NewService(Config{Origins: nil}, logger)
				require.Error(t, err)
				require.Nil(t, service)
			})
		})
		t.Run("Success", func(t *testing.T) {
			service, err := NewService(Config{
				Origins: []string{"example.com", "example.fr", "www.negrel.dev", "*.github.com"},
			}, logger)
			require.NoError(t, err)
			require.NotNil(t, service)
		})
	})

	t.Run("IsOriginRegistered", func(t *testing.T) {
		ctx := context.Background()

		t.Run("NonRegistered", func(t *testing.T) {
			service, err := NewService(Config{Origins: []string{"notexample.com"}}, logger)
			require.NoError(t, err)

			isRegistered, err := service.IsOriginRegistered(ctx, "example.com")
			require.NoError(t, err)
			require.False(t, isRegistered)
		})

		t.Run("Registered", func(t *testing.T) {
			service, err := NewService(Config{Origins: []string{"example.org", "example.com", "example.io", "*.example.fr"}}, logger)
			require.NoError(t, err)

			isRegistered, err := service.IsOriginRegistered(ctx, "example.com")
			require.NoError(t, err)
			require.True(t, isRegistered)

			isRegistered, err = service.IsOriginRegistered(ctx, "example.org")
			require.NoError(t, err)
			require.True(t, isRegistered)

			isRegistered, err = service.IsOriginRegistered(ctx, "example.io")
			require.NoError(t, err)
			require.True(t, isRegistered)

			isRegistered, err = service.IsOriginRegistered(ctx, "example.fr")
			require.NoError(t, err)
			require.False(t, isRegistered)

			isRegistered, err = service.IsOriginRegistered(ctx, "foo.bar.baz.example.fr")
			require.NoError(t, err)
			require.True(t, isRegistered)
		})
	})
}

func BenchmarkServiceNoWildcard(b *testing.B) {
	logger := log.New("env_var_service_test", io.Discard, false)
	service, err := NewService(Config{Origins: []string{"example.org", "example.com", "example.io"}}, logger)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	origins := []string{"example.com", "foo.example.com", "www.negrel.dev"}
	for i := range b.N {
		_, _ = service.IsOriginRegistered(ctx, origins[i%len(origins)])
	}
}

func BenchmarkServiceWithWildcards(b *testing.B) {
	logger := log.New("env_var_service_test", io.Discard, false)
	service, err := NewService(Config{Origins: []string{"example.org", "example.com", "*.example.io"}}, logger)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	origins := []string{"example.com", "foo.bar.example.io", "www.negrel.dev"}
	for i := range b.N {
		_, _ = service.IsOriginRegistered(ctx, origins[i%len(origins)])
	}
}
