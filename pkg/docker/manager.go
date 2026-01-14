package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// manager handles docker container operations
type Manager struct {
	cli *client.Client
}

// newmanager creates a new docker manager
func NewManager() (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Manager{cli: cli}, nil
}

// startcontainer starts a container for a given function image
// returns the container ip and id
func (m *Manager) StartContainer(ctx context.Context, imageTag string, name string) (string, string, error) {
	// check if image exists locally
	_, _, err := m.cli.ImageInspectWithRaw(ctx, imageTag)
	if client.IsErrNotFound(err) {
		// attempt to pull (assuming it's in a registry, though for this local demo it might be built locally)
		// for local dev without a registry, we skip pull if it's not found and fail.
		// in a real scenario, we'd pull:
		reader, err := m.cli.ImagePull(ctx, imageTag, types.ImagePullOptions{})
		if err != nil {
			return "", "", fmt.Errorf("failed to pull image %s: %w", imageTag, err)
		}
		defer reader.Close()
		io.Copy(io.Discard, reader) // wait for pull to finish
	}

	// create container
	// we bind to a random host port to avoid conflicts, or use internal network
	// for this demo, we'll let docker assign a random host port
	config := &container.Config{
		Image: imageTag,
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}
	
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "0", // random available port
				},
			},
		},
		AutoRemove: true, // clean up after stop
	}

	networkConfig := &network.NetworkingConfig{}

	resp, err := m.cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, "nanolambda-"+name+"-"+fmt.Sprintf("%d", time.Now().UnixNano()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create container: %w", err)
	}

	// start container
	if err := m.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", fmt.Errorf("failed to start container: %w", err)
	}

	// inspect to get the assigned port
	inspect, err := m.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect container: %w", err)
	}

	// get the host port that maps to 8080
	ports := inspect.NetworkSettings.Ports["8080/tcp"]
	if len(ports) == 0 {
		return "", "", fmt.Errorf("no ports bound")
	}
	
	hostPort := ports[0].HostPort
	hostIP := "127.0.0.1" // localhost for this architecture

	return hostIP + ":" + hostPort, resp.ID, nil
}

// stopcontainer kills a container
func (m *Manager) StopContainer(ctx context.Context, containerID string) error {
	return m.cli.ContainerStop(ctx, containerID, container.StopOptions{})
}

// listrunningcontainers returns a list of containers managed by nanolambda
func (m *Manager) ListRunningContainers(ctx context.Context) ([]types.Container, error) {
	return m.cli.ContainerList(ctx, types.ContainerListOptions{
		// filter by name prefix if desired, or labels
		All: false,
	})
}