package main

import (
	"context"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	registryPrefix         = "ghcr.io/nordsecurity/nordvpn-linux/"
	imageBuilder           = registryPrefix + "builder:1.1.1"
	imagePackager          = registryPrefix + "packager:1.0.1"
	imageProtobufGenerator = registryPrefix + "generator:1.0.1"
	imageScanner           = registryPrefix + "scanner:1.0.0"
	imageTester            = registryPrefix + "tester:1.0.2"
	imageQAPeer            = registryPrefix + "qa-peer:1.0.2"
	imageLinter            = registryPrefix + "linter:1.0.0"
	imageRuster            = registryPrefix + "ruster:1.0.1"
)

// Build is used for native builds.
type Build mg.Namespace

// Generate is used for code generation.
type Generate mg.Namespace

// Install is used for putting binary in specific locations.
type Install mg.Namespace

// Run is used to run our application in specific environments
type Run mg.Namespace

// Test is used for various testing scenarios.
type Test mg.Namespace

// View is used for viewing information about the project.
type View mg.Namespace

// Coverage for pure Go
func (View) Coverage() error {
	return sh.Run("go", "tool", "cover", "-html=coverage.txt")
}

// Docs for the nordvpn application
func (View) Docs() error {
	fmt.Println("Open http://localhost:6060/pkg/nordvpn to view documentation")
	return sh.Run("godoc")
}

// Download external dependencies
func Download() error {
	env, err := getEnv()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	env["ARCH"] = build.Default.GOARCH
	env["CI_PROJECT_DIR"] = cwd
	return sh.RunWith(env, "ci/check_dependencies.sh")
}

// Data for Linux packages
func (Build) Data() error {
	if internal.FileExists("dist/data") {
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env := map[string]string{"CI_PROJECT_DIR": cwd}
	return sh.RunWith(env, "ci/data.sh")
}

// Notices for third party dependencies
func (Build) Notices() error {
	if internal.FileExists("dist/THIRD-PARTY-NOTICES.md") {
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env := map[string]string{"CI_PROJECT_DIR": cwd}
	return sh.RunWith(env, "ci/licenses.sh")
}

func buildPackage(packageType string, buildFlags string) error {
	mg.Deps(Build.Data)
	mg.Deps(mg.F(buildBinaries, buildFlags))
	mg.Deps(Build.Openvpn)
	mg.Deps(Build.Notices)

	env, err := getEnv()
	if err != nil {
		return err
	}
	env["ARCH"] = build.Default.GOARCH
	env["GOPATH"] = build.Default.GOPATH
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env["CI_PROJECT_DIR"] = cwd
	return sh.RunWith(env, "ci/nfpm/build_packages_resources.sh", packageType)
}

// Deb package for the host architecture
func (Build) Deb() error {
	return buildPackage("deb", "")
}

// Rpm package for the host architecture
func (Build) Rpm() error {
	return buildPackage("rpm", "")
}

func buildPackageDocker(ctx context.Context, packageType string, buildFlags string) error {
	mg.Deps(Build.Data)
	mg.Deps(mg.F(buildBinariesDocker, buildFlags))
	mg.Deps(Build.OpenvpnDocker)
	mg.Deps(Build.Notices)

	env, err := getEnv()
	if err != nil {
		return err
	}

	git, err := getGitInfo()
	if err != nil {
		return err
	}

	env["ARCH"] = build.Default.GOARCH
	env["CI_PROJECT_DIR"] = "/opt"
	env["ENVIRONMENT"] = "dev"
	env["HASH"] = git.commitHash
	env["PACKAGE"] = "source"
	env["VERSION"] = git.versionTag
	return RunDocker(
		ctx,
		env,
		imagePackager,
		[]string{"ci/nfpm/build_packages_resources.sh", packageType},
	)
}

// DebDocker package using Docker builder
func (Build) DebDocker(ctx context.Context) error {
	return buildPackageDocker(ctx, "deb", "")
}

// RpmDocker package using Docker builder
func (Build) RpmDocker(ctx context.Context) error {
	return buildPackageDocker(ctx, "rpm", "")
}

func buildBinaries(buildFlags string) error {
	mg.Deps(Download)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	env, err := getEnv()
	if err != nil {
		return err
	}

	if !strings.Contains(env["FEATURES"], "internal") {
		mg.Deps(Build.Rust)
	}

	git, err := getGitInfo()
	if err != nil {
		return err
	}
	env["ARCH"] = build.Default.GOARCH
	env["CI_PROJECT_DIR"] = cwd
	env["HASH"] = git.commitHash
	env["PACKAGE"] = "source"
	env["VERSION"] = git.versionTag
	env["ENVIRONMENT"] = "dev"
	env["BUILD_FLAGS"] = buildFlags

	return sh.RunWith(env, "ci/compile.sh")
}

// Binaries from cmd/* for the host architecture
func (Build) Binaries() error {
	return buildBinaries("")
}

func buildBinariesDocker(ctx context.Context, buildFlags string) error {
	mg.Deps(Download)

	env, err := getEnv()
	if err != nil {
		return err
	}

	if !strings.Contains(env["FEATURES"], "internal") {
		mg.Deps(Build.RustDocker)
	}

	git, err := getGitInfo()
	if err != nil {
		return err
	}
	env["ARCH"] = build.Default.GOARCH
	env["CI_PROJECT_DIR"] = "/opt"
	env["ENVIRONMENT"] = "dev"
	env["HASH"] = git.commitHash
	env["PACKAGE"] = "source"
	env["VERSION"] = git.versionTag
	env["BUILD_FLAGS"] = buildFlags

	return RunDocker(
		ctx,
		env,
		imageBuilder,
		[]string{"ci/compile.sh"},
	)
}

// Builds all binaries using Docker builder
func (Build) BinariesDocker(ctx context.Context) error {
	return buildBinariesDocker(ctx, "")
}

// Openvpn binaries for the host architecture
func (Build) Openvpn(ctx context.Context) error {
	mg.Deps(Download)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env, err := getEnv()
	if err != nil {
		return err
	}

	env["ARCH"] = "amd64"
	env["CI_PROJECT_DIR"] = cwd

	return sh.RunWith(env, "build/openvpn/build.sh")
}

// Openvpn binaries for the host architecture
func (Build) OpenvpnDocker(ctx context.Context) error {
	mg.Deps(Download)

	env, err := getEnv()
	if err != nil {
		return err
	}

	env["ARCH"] = "amd64"
	env["CI_PROJECT_DIR"] = "/opt"
	return RunDocker(
		ctx,
		env,
		imageBuilder,
		[]string{"build/openvpn/build.sh"},
	)
}

// Rust dependencies for the host architecture
func (Build) Rust(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env := map[string]string{
		"ARCHS":          build.Default.GOARCH,
		"CI_PROJECT_DIR": cwd,
	}
	return sh.RunWith(env, "build/foss/build.sh")
}

// Builds rust dependencies using Docker builder
func (Build) RustDocker(ctx context.Context) error {
	env, err := getEnv()
	if err != nil {
		return err
	}

	env["ARCHS"] = "amd64"
	env["CI_PROJECT_DIR"] = "/opt"
	if err := RunDocker(
		ctx,
		env,
		imageRuster,
		[]string{"build/foss/build.sh"},
	); err != nil {
		return err
	}

	return nil
}

// Generate Protobuf from protobuf/* definitions using Docker builder
func (Generate) ProtobufDocker(ctx context.Context) error {
	mg.Deps(Download)

	env, err := getEnv()
	if err != nil {
		return err
	}

	return RunDocker(
		ctx,
		env,
		imageProtobufGenerator,
		[]string{"ci/generate_protobuf.sh"},
	)
}

// run unit tests
func (Test) Go() error {
	if err := sh.Rm("coverage.txt"); err != nil {
		return err
	}

	env, err := getEnv()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env["CI_PROJECT_DIR"] = cwd

	return sh.RunWithV(env, "ci/test.sh")
}

// run cgo tests
func (Test) CgoDocker(ctx context.Context) error {
	if err := sh.Rm("coverage.txt"); err != nil {
		return err
	}

	env, err := getEnv()
	if err != nil {
		return err
	}
	env["ARCH"] = build.Default.GOARCH
	env["CI_PROJECT_DIR"] = "/opt"
	env["ENVIRONMENT"] = "dev"
	env["GOPATH"] = build.Default.GOPATH

	return RunDockerWithSettings(
		ctx,
		env,
		imageBuilder,
		[]string{"ci/test.sh", "full"},
		DockerSettings{Privileged: true},
	)
}

// Gosec run security scan on the host
func (Test) Gosec() error {
	mg.Deps(Install.Gosec)
	return sh.RunV("ci/scan.sh")
}

// GosecDocker run security scan inside of docker container
func (Test) GosecDocker(ctx context.Context) error {
	env, err := getEnv()
	if err != nil {
		return err
	}
	env["CI_PROJECT_DIR"] = "/opt"

	return RunDocker(
		ctx,
		env,
		imageScanner,
		[]string{"ci/scan.sh"},
	)
}

// Hardening test ELF binaries
func (Test) Hardening(ctx context.Context) error {
	env, err := getEnv()
	if err != nil {
		return err
	}
	env["ARCH"] = build.Default.GOARCH
	env["CI_PROJECT_DIR"] = "/opt"

	return RunDocker(
		ctx,
		env,
		imageScanner,
		[]string{"ci/hardening.sh"},
	)
}

// Run QA tests in Docker container (arguments: {testGroup} {testPattern})
func (Test) QADocker(ctx context.Context, testGroup, testPattern string) error {
	mg.Deps(mg.F(buildPackageDocker, "deb", "-cover"))
	return qa(ctx, testGroup, testPattern)
}

// Run QA tests in Docker container, builds the package locally and skips the build if it is already
// present in the package directory (arguments: {testGroup} {testPattern})
func (Test) QADockerFast(ctx context.Context, testGroup, testPattern string) error {
	const debPath string = "dist/app/deb/nordvpn_*_amd64.deb"
	matches, err := filepath.Glob(debPath)

	if len(matches) == 0 || err != nil {
		mg.Deps(mg.F(buildPackage, "deb", "-cover"))
	}

	return qa(ctx, testGroup, testPattern)
}

func qa(ctx context.Context, testGroup, testPattern string) error {
	env, err := getEnv()
	if err != nil {
		return err
	}
	env["CI_PROJECT_DIR"] = "/opt"
	env["QA_PEER_ADDRESS"] = "http://qa-peer:8000/exec"
	env["COVERDIR"] = "covdatafiles"

	dir := env["CI_PROJECT_DIR"] + "/" + env["COVERDIR"]
	_ = os.RemoveAll(dir)

	_ = RemoveDockerNetwork(context.Background(), "qa") // Needed if job was killed
	networkID, err := CreateDockerNetwork(ctx, "qa")
	if err != nil {
		return fmt.Errorf("%w (while creating network)", err)
	}

	containerStoppedChan := make(chan interface{})

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		<-containerStoppedChan
		err = RemoveDockerNetwork(context.Background(), networkID)
		if err != nil {
			fmt.Println(err)
		}
	}()

	err = RunDockerWithSettings(ctx, env,
		imageQAPeer,
		[]string{},
		DockerSettings{Privileged: true, Daemonize: true, Network: networkID, DaemonizeStopChan: containerStoppedChan},
	)
	if err != nil {
		return fmt.Errorf("%w (while starting qa-peer)", err)
	}

	return RunDockerWithSettings(ctx, env,
		imageTester,
		[]string{"ci/test_deb.sh", testGroup, testPattern},
		DockerSettings{Privileged: true, Network: networkID},
	)
}

// Performs linter check against Go codebase
func (Test) Lint() error {
	mg.Deps(Download)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	env := map[string]string{
		"CI_PROJECT_DIR": cwd,
		"ENVIRONMENT":    "dev",
		"ARCH":           build.Default.GOARCH,
	}
	return sh.RunWithV(env, "ci/lint.sh")
}

// Performs linter check against Go codebase in Docker container
func (Test) LintDocker(ctx context.Context) error {
	mg.Deps(Download)

	env, err := getEnv()
	if err != nil {
		return err
	}

	env["ARCH"] = build.Default.GOARCH
	env["ENVIRONMENT"] = "dev"
	env["CI_PROJECT_DIR"] = "/opt"

	return RunDocker(
		ctx,
		env,
		imageLinter,
		[]string{"ci/lint.sh"},
	)
}

// Binaries to their respective locations and restart
// the daemon along the way
func (Install) Binaries() error {
	mg.Deps(Build.Binaries)
	systemctl := sh.RunCmd("sudo", "systemctl")
	if err := systemctl("stop", "nordvpnd"); err != nil {
		return err
	}

	cp := sh.RunCmd("sudo", "cp", "--preserve")
	nordvpnDst, err := sh.Output("which", "nordvpn")
	if err != nil {
		nordvpnDst = "/usr/bin/nordvpn"
	}

	nordvpnSrc := fmt.Sprintf("bin/%s/nordvpn", build.Default.GOARCH)
	if err := cp(nordvpnSrc, nordvpnDst); err != nil {
		return err
	}

	nordvpndDst, err := sh.Output("which", "nordvpnd")
	if err != nil {
		nordvpndDst = "/usr/sbin/nordvpnd"
	}

	nordvpndSrc := fmt.Sprintf("bin/%s/nordvpnd", build.Default.GOARCH)
	if err := cp(nordvpndSrc, nordvpndDst); err != nil {
		return err
	}

	filesharedDst, err := sh.Output("which", internal.Fileshared)
	if err != nil {
		filesharedDst = "/usr/bin/" + internal.Fileshared
	}

	filesharedSrc := fmt.Sprintf("bin/%s/%s", build.Default.GOARCH, internal.Fileshared)
	if err := cp(filesharedSrc, filesharedDst); err != nil {
		return err
	}

	return systemctl("start", "nordvpnd")
}

// Gosec source code scanner to ~/go/bin
func (Install) Gosec() error {
	installPath := build.Default.GOPATH + "/bin"
	if internal.FileExists(installPath + "/gosec") {
		return nil
	}

	file, err := os.CreateTemp("", "gosec")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	script, err := sh.Output(
		"curl",
		"-sfL",
		"https://raw.githubusercontent.com/securego/gosec/master/install.sh",
	)
	if err != nil {
		return err
	}

	if _, err := file.Write([]byte(script)); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return sh.RunV("sh", file.Name(), "-b", installPath, "v2.12.0")
}

// Docker builds and runs nordvpn app in Docker container
func (Run) Docker() error {
	mg.Deps(Build.DebDocker)
	return docker()
}

// DockerFast builds and runs nordvpn app in Docker container
func (Run) DockerFast() error {
	mg.Deps(Build.Deb)
	return docker()
}

func docker() error {
	tag := "mage-nordvpn"
	err := BuildDocker("ci/docker/nordvpn/Dockerfile.dev", tag)
	if err != nil {
		return err
	}

	env, err := getEnv()
	if err != nil {
		return err
	}

	// RunDocker has problems with allowing interactivity, so running it this way
	// #nosec G204 -- used only during development/testing
	cmd := exec.Command("docker", "run", "-e", "NORDVPN_LOGIN_TOKEN="+env["DEFAULT_LOGIN_TOKEN"],
		"--cap-add=NET_ADMIN", "-it", "--rm", tag)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
