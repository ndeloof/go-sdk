package container_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"

	"github.com/docker/go-sdk/container"
)

func TestContainer_Attach(t *testing.T) {
	t.Run("running-container", func(t *testing.T) {
		ctr, err := container.Run(context.Background(),
			// using an image that has a long-running command
			container.WithImage(alpineLatest),
			container.WithCmd("echo", "hello world"),
			container.WithNoStart(),
			container.WithAttachStdout(),
			// container.WithTTY()
		)
		require.NoError(t, err)

		container.Cleanup(t, ctr)

		t.Run("without TTY", func(t *testing.T) {
			_, multiplexed, err := ctr.Attach(context.Background(), container.WithStdout, container.WithStderr)
			require.NoError(t, err)
			require.NotNil(t, multiplexed)

			err = ctr.Start(context.Background())
			require.NoError(t, err)

			var stdout, stderr bytes.Buffer
			_, err = stdcopy.StdCopy(&stdout, &stderr, multiplexed)
			require.NoError(t, err)
			require.Equal(t, "hello world\n", stdout.String())
		})
	})
}
