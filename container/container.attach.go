package container

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type attachOptions func(options container.AttachOptions)

// WithStdin attach to container stdin
func WithStdin(options container.AttachOptions) {
	options.Stdin = true
}

// WithStdout attach to container stdout
func WithStdout(options container.AttachOptions) {
	options.Stdout = true
}

// WithStderr attach to container stderr
func WithStderr(options container.AttachOptions) {
	options.Stderr = true
}

// Attach connects to container's std(in|out|err) streams so one can interact with container
// It returns an [io.Writer] connected to stdin, an [io.Reader] containing the combined stdout and stderr,
// and any encountered error. Note that reading directly from the [io.Reader]
// may result in unexpected bytes due to custom stream multiplexing headers if container runs without a tty.
// [github.com/docker/docker/pkg/stdcopy.StdCopy] from the Docker API should then be used.
func (c *Container) Attach(ctx context.Context, options ...attachOptions) (io.WriteCloser, io.ReadCloser, error) {
	o := container.AttachOptions{Stream: true}
	for _, opt := range options {
		opt(o)
	}

	cnx, err := c.dockerClient.ContainerAttach(ctx, c.ID(), o)
	if err != nil {
		return nil, nil, fmt.Errorf("container attach: %w", err)
	}
	return ContainerStdin{HijackedResponse: cnx}, ContainerStdout{HijackedResponse: cnx}, nil
}

var _ io.ReadCloser = ContainerStdout{}

// ContainerStdout implement ReadCloser for moby.HijackedResponse
type ContainerStdout struct {
	types.HijackedResponse
}

// Read implement io.ReadCloser
func (l ContainerStdout) Read(p []byte) (n int, err error) {
	return l.Reader.Read(p)
}

// Close implement io.ReadCloser
func (l ContainerStdout) Close() error {
	l.HijackedResponse.Close()
	return nil
}

var _ io.WriteCloser = ContainerStdin{}

// ContainerStdin implement WriteCloser for moby.HijackedResponse
type ContainerStdin struct {
	types.HijackedResponse
}

// Write implement io.WriteCloser
func (c ContainerStdin) Write(p []byte) (n int, err error) {
	return c.Conn.Write(p)
}

// Close implement io.WriteCloser
func (c ContainerStdin) Close() error {
	return c.CloseWrite()
}
