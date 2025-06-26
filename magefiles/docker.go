//go:build mage

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/exp/slices"
)

type DockerSettings struct {
	Privileged        bool
	Daemonize         bool       // Container is run in background, output is not available then
	DaemonizeStopChan chan<- any // Value will be sent on this channel when daemonized docker container is stopped
	Network           string
}

func RunDocker(
	ctx context.Context,
	env map[string]string,
	image string,
	cmd []string,
) error {
	settings := DockerSettings{}
	return runDocker(ctx, env, image, cmd, settings.Privileged, settings.Daemonize, nil, settings.Network)
}

func RunDockerWithSettings(
	ctx context.Context,
	env map[string]string,
	image string,
	cmd []string,
	settings DockerSettings,
) error {
	return runDocker(ctx, env, image, cmd, settings.Privileged, settings.Daemonize, settings.DaemonizeStopChan, settings.Network)
}

// BuildDocker image with project root as context
func BuildDocker(dockerfile, tag string) error {
	// Golang docker integration doesn't seem to support passing context via path,
	// and workarounds seem complicated, so this will probably be easier to maintain
	// https://stackoverflow.com/questions/38804313/build-docker-image-from-go-code

	fmt.Println("building docker image")
	cmd := exec.Command("docker", "build", "-f", dockerfile, "-t", tag, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runDocker(
	ctx context.Context,
	env map[string]string,
	img string,
	cmd []string,
	isPrivileged bool,
	daemonize bool,
	containerStoppedChan chan<- any,
	network string,
) error {
	fmt.Println("creating docker client")
	defer func() {
		if !isPrivileged {
			return
		}
		if err := fixPermissions(); err != nil {
			fmt.Println("failed to fix permissions", err)
		}
	}()
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer docker.Close()

	pullImage := true
	if idempotent, ok := env["IDEMPOTENT_DOCKER"]; ok && idempotent == "1" {
		list, err := docker.ImageList(context.Background(), types.ImageListOptions{})
		if err != nil {
			return err
		}

		imageIndex := slices.IndexFunc(list, func(imageSummary image.Summary) bool {
			tagIndex := slices.Index(imageSummary.RepoTags, img)
			return tagIndex != -1
		})
		pullImage = imageIndex == -1
	}

	if pullImage {
		fmt.Printf("pulling docker image %s\n", img)
		if err := pullDocker(ctx, docker, img); err != nil {
			return err
		}
	}

	name := getNameFromImage(img)
	fmt.Printf("creating %s docker container\n", name)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: cwd,
			Target: "/opt",
		},
		{
			Type:   mount.TypeBind,
			Source: "/lib/modules",
			Target: "/lib/modules",
		},
	}

	user := ""
	if !isPrivileged {
		user = fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
		env["HOME"] = "/home/user"
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeTmpfs,
			Target: "/home/user",
		})
	}

	containerConfig := container.Config{
		Image:      img,
		Cmd:        cmd,
		Env:        envMapToList(env),
		WorkingDir: "/opt",
		User:       user,
	}

	if env["MOUNT_HOST_GOMODCACHE"] == "1" {
		goModCache, err := getGoModCache()
		if err != nil {
			fmt.Println("Error on retrieving go mod cache on host:", err)
		} else {
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: goModCache,
				Target: "/go/pkg/mod",
			})
		}
	}

	hostConfig := container.HostConfig{
		AutoRemove:  true,
		Mounts:      mounts,
		Privileged:  isPrivileged,
		Sysctls:     map[string]string{"net.ipv6.conf.all.disable_ipv6": "1"},
		NetworkMode: container.NetworkMode(network),
	}

	resp, err := docker.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, name)
	if err != nil {
		return err
	}

	fmt.Printf("starting %s docker container\n", name)
	if err := docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	if daemonize {
		go func() {
			<-ctx.Done()
			timeoutSec := 1
			err := docker.ContainerStop(context.Background(), resp.ID, container.StopOptions{Timeout: &timeoutSec})
			if containerStoppedChan != nil {
				containerStoppedChan <- true
			}
			if err != nil {
				fmt.Println(err)
			}
		}()
		return nil
	}

	fmt.Printf("waiting for %s docker container to finish\n", name)
	statusCh, errCh := docker.ContainerWait(ctx, resp.ID, container.WaitConditionRemoved)
	attach, err := docker.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return err
	}
	defer attach.Close()

	for {
		select {
		case err := <-errCh:
			return err
		case status := <-statusCh:
			if status.StatusCode != 0 {
				return fmt.Errorf("exit code %d", status.StatusCode)
			}
			return nil
		default:
			_, err := stdcopy.StdCopy(os.Stdout, os.Stderr, attach.Reader)
			if err != nil {
				return err
			}
		}
	}
}

func fixPermissions() error {
	fmt.Println("fixing permissions")
	user := os.Getenv("USER")
	if user == "" {
		return fmt.Errorf("USER environment variable not set")
	}

	ug := fmt.Sprintf("%s:%s", user, user)
	cmd := exec.Command("sudo", "chown", "-R", ug, ".")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("Permissions updated successfully.")
	return nil
}

func pullDocker(
	ctx context.Context,
	cli *client.Client,
	image string,
) error {
	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pulling %s: %w", image, err)
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader) // otherwise exits without finishing pulling
	return err
}

func envMapToList(env map[string]string) []string {
	envList := []string{}
	for key, value := range env {
		envList = append(envList, fmt.Sprintf("%s=%s", key, value))
	}
	return envList
}

// Extracts "tester" from "ghcr.io/nordsecurity/nordvpn-linux/tester:1.0.0"
func getNameFromImage(image string) string {
	imageSplit := strings.Split(image, "/")
	return strings.Split(imageSplit[len(imageSplit)-1], ":")[0]
}

// CreateDockerNetwork returns ID of newly created network
func CreateDockerNetwork(ctx context.Context, name string) (string, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}

	resp, err := docker.NetworkCreate(ctx, name, types.NetworkCreate{})
	return resp.ID, err
}

// RemoveDockerNetwork using ID returned from CreateDockerNetwork
func RemoveDockerNetwork(ctx context.Context, id string) error {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	return docker.NetworkRemove(ctx, id)
}

func getGoModCache() (string, error) {
	// `GOMODCACHE` is not accessible from go code directly during build, therefore it
	// has to be acquired in a separate process.
	out, err := exec.Command("go", "env", "GOMODCACHE").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("executing 'go env GOMODCACHE': %w", err)
	}
	goModCache := strings.TrimSpace(string(out))
	if goModCache == "" {
		return "", fmt.Errorf("GOMODCACHE is empty")
	}
	return goModCache, nil
}
