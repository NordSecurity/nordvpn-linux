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
	imageBuilder           = registryPrefix + "builder:1.1.5"
	imagePackager          = registryPrefix + "packager:1.0.3"
	imageSnapPackager      = registryPrefix + "snaper:0.0.2"
	imageProtobufGenerator = registryPrefix + "generator:1.0.2"
	imageScanner           = registryPrefix + "scanner:1.0.3"
	imageTester            = registryPrefix + "tester:1.1.10"
	imageQAPeer            = registryPrefix + "qa-peer:1.0.4"
	imageRuster            = registryPrefix + "ruster:1.0.4"

	dockerWorkDir  = "/opt"
	devPackageType = "source"
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

// installHookIfNordsec installs git-secrets hook if git user belongs to Nord Security.
func installHookIfNordsec() error {
	output, err := exec.Command("git", "config", "--get", "user.email").CombinedOutput()
	if err != nil {
		if werr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 is returned when user.email is not configured. In this case we want to
			// skip the validation in order to allow users to build without configuring git environment.
			if werr.ExitCode() != 1 {
				return fmt.Errorf("getting user.email from git config: %w", err)
			}
		} else {
			return fmt.Errorf("unknown error when getting user.email from git config: %w", err)
		}

		fmt.Println("Warning: git user.email is not configured.")
	}

	if !strings.Contains(string(output), "nordsec") {
		return nil
	}

	output, err = exec.Command("git", "secrets", "--list").CombinedOutput()
	if err != nil {
		return fmt.Errorf("fetching git-secrets providers: %w", err)
	}

	if !strings.Contains(string(output), "llt-secrets") {
		return fmt.Errorf("Secret provider was not configured.")
	}

	if _, err := exec.Command("git", "secrets", "--install", "--force").CombinedOutput(); err != nil {
		return fmt.Errorf("installing git-secrets hook: %w", err)
	}

	return nil
}

// Coverage for pure Go
func (View) Coverage() error {
	return sh.Run("go", "tool", "cover", "-html=coverage.txt")
}

// Docs for the nordvpn application
func (View) Docs() error {
	fmt.Println("Open http://localhost:6060/pkg/nordvpn to view documentation")
	return sh.Run("godoc")
}

// Clean is used to clean build results.
func Clean() error {
	// cleanup regular build folders
	buildFolders := []string{"./bin", "./dist"}
	for _, folder := range buildFolders {
		if internal.FileExists(folder) {
			fmt.Println("Cleanup build folder:", folder)
			if err := sh.Run("rm", "-r", folder); err != nil {
				return err
			}
		}
	}
	// cleanup folders left after building snap in docker
	// if folders do not exist - no problem, no error
	snapFolders := []string{"./parts", "./stage", "./prime"}
	for _, folder := range snapFolders {
		if internal.FileExists(folder) {
			fmt.Println("Cleanup snapcraft folder:", folder)
			if err := sh.Run("sudo", "rm", "-r", folder); err != nil {
				return err
			}
		}
	}
	// cleanup snap packages in current dir
	pattern := "*.snap"
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		fmt.Println("Cleanup snaps...")
		for _, snap := range matches {
			fmt.Println("Cleanup snap:", snap)
			// sudo is needed when snap is built using docker
			if err := sh.Run("sudo", "rm", snap); err != nil {
				return err
			}
		}
	}
	// not everybody have/use snapcraft
	if err := sh.Run("which", "snapcraft"); err == nil {
		fmt.Println("Cleanup snapcraft internals...")
		if err := sh.Run("snapcraft", "clean"); err != nil {
			return err
		}
	}
	return nil
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
	env["WORKDIR"] = cwd
	return sh.RunWith(env, "ci/check_dependencies.sh")
}

// Download OpenVPN external dependencies
func DownloadOpenvpn() error {
	env, err := getEnv()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	env["ARCH"] = build.Default.GOARCH
	env["WORKDIR"] = cwd
	return sh.RunWith(env, "build/openvpn/check_dependencies.sh")
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
	env := map[string]string{"WORKDIR": cwd}
	env["GOPATH"] = build.Default.GOPATH
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
	env := map[string]string{"WORKDIR": cwd}
	return sh.RunWith(env, "ci/licenses.sh")
}

func buildPackage(packageType string, buildFlags string) error {
	mg.Deps(Build.Data)
	mg.Deps(mg.F(buildBinaries, buildFlags))
	mg.Deps(Build.Notices)

	// do not build openvpn dependency if it already exists
	if !internal.FileExists(fmt.Sprintf("./bin/deps/openvpn/%s/latest/openvpn", build.Default.GOARCH)) {
		mg.Deps(Build.Openvpn)
	}

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
	env["WORKDIR"] = cwd
	if packageType == "snap" {
		return sh.RunWith(env, "snapcraft")
	}
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

// Snap package for the host architecture
func (Build) Snap() error {
	return buildPackage("snap", "")
}

func buildPackageDocker(ctx context.Context, packageType string, buildFlags string) error {
	mg.Deps(Build.Data)
	mg.Deps(mg.F(buildBinariesDocker, buildFlags))
	mg.Deps(Build.Notices)

	// do not build openvpn dependency if it already exists
	if !internal.FileExists(fmt.Sprintf("./bin/deps/openvpn/%s/latest/openvpn", build.Default.GOARCH)) {
		mg.Deps(Build.OpenvpnDocker)
	}

	env, err := getEnv()
	if err != nil {
		return err
	}

	git, err := getGitInfo()
	if err != nil {
		return err
	}

	env["ARCH"] = build.Default.GOARCH
	env["WORKDIR"] = dockerWorkDir
	env["ENVIRONMENT"] = string(internal.Development)
	env["HASH"] = git.commitHash
	env["PACKAGE"] = devPackageType
	env["VERSION"] = git.versionTag
	if packageType == "snap" {
		return RunDocker(
			ctx,
			env,
			imageSnapPackager,
			[]string{"snapcraft", "--destructive-mode"},
		)
	}
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

// SnapDocker package using Docker builder
func (Build) SnapDocker(ctx context.Context) error {
	return buildPackageDocker(ctx, "snap", "")
}

func buildBinaries(buildFlags string) error {
	if err := installHookIfNordsec(); err != nil {
		return err
	}

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
	env["WORKDIR"] = cwd
	env["HASH"] = git.commitHash
	env["PACKAGE"] = devPackageType
	env["VERSION"] = git.versionTag
	env["ENVIRONMENT"] = string(internal.Development)
	env["BUILD_FLAGS"] = buildFlags

	return sh.RunWith(env, "ci/compile.sh")
}

// Binaries from cmd/* for the host architecture
func (Build) Binaries() error {
	return buildBinaries("")
}

func buildBinariesDocker(ctx context.Context, buildFlags string) error {
	if err := installHookIfNordsec(); err != nil {
		return err
	}

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
	env["WORKDIR"] = dockerWorkDir
	env["ENVIRONMENT"] = string(internal.Development)
	env["HASH"] = git.commitHash
	env["PACKAGE"] = devPackageType
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
	mg.Deps(DownloadOpenvpn)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env, err := getEnv()
	if err != nil {
		return err
	}

	env["ARCH"] = build.Default.GOARCH
	env["WORKDIR"] = cwd

	return sh.RunWith(env, "build/openvpn/build.sh")
}

// Openvpn binaries for the host architecture
func (Build) OpenvpnDocker(ctx context.Context) error {
	mg.Deps(DownloadOpenvpn)

	env, err := getEnv()
	if err != nil {
		return err
	}

	env["ARCH"] = build.Default.GOARCH
	env["WORKDIR"] = dockerWorkDir
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
		"ARCHS":   build.Default.GOARCH,
		"WORKDIR": cwd,
	}
	return sh.RunWith(env, "build/foss/build.sh")
}

// Builds rust dependencies using Docker builder
func (Build) RustDocker(ctx context.Context) error {
	env, err := getEnv()
	if err != nil {
		return err
	}

	env["ARCHS"] = build.Default.GOARCH
	env["WORKDIR"] = dockerWorkDir
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
	env["WORKDIR"] = cwd

	return sh.RunWithV(env, "ci/test.sh")
}

// run cgo tests
func (Test) CgoDocker(ctx context.Context) error {
	mg.Deps(Download)
	if err := sh.Rm("coverage.txt"); err != nil {
		return err
	}

	env, err := getEnv()
	if err != nil {
		return err
	}
	env["ARCH"] = build.Default.GOARCH
	env["WORKDIR"] = dockerWorkDir
	env["ENVIRONMENT"] = string(internal.Development)

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
	return sh.Run("gosec", "-quiet", "-exclude-dir=third-party", "./...")
}

// Hardening test ELF binaries
func (Test) Hardening(ctx context.Context) error {
	env, err := getEnv()
	if err != nil {
		return err
	}
	env["ARCH"] = build.Default.GOARCH
	env["WORKDIR"] = dockerWorkDir

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
	debPath := fmt.Sprintf("dist/app/deb/nordvpn_*%s.deb", build.Default.GOARCH)
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
	env["WORKDIR"] = dockerWorkDir
	env["QA_PEER_ADDRESS"] = "http://qa-peer:8000/exec"
	env["COVERDIR"] = "covdatafiles"

	dir := env["WORKDIR"] + "/" + env["COVERDIR"]
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
	return sh.Run("golangci-lint", "run", "-v", "--config=.golangci-lint.yml")
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

	filesharedDst := internal.FileshareBinaryPath

	filesharedSrc := fmt.Sprintf("bin/%s/%s", build.Default.GOARCH, internal.Fileshare)
	if err := cp(filesharedSrc, filesharedDst); err != nil {
		return err
	}

	norduserDst, err := sh.Output("which", internal.Norduserd)
	if err != nil {
		norduserDst = fmt.Sprintf("/usr/bin/%s", internal.Norduserd)
	}

	norduserSrc := fmt.Sprintf("bin/%s/%s", build.Default.GOARCH, internal.Norduserd)
	if err := cp(norduserSrc, norduserDst); err != nil {
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
	debPath := fmt.Sprintf("dist/app/deb/nordvpn_*%s.deb", build.Default.GOARCH)
	matches, err := filepath.Glob(debPath)
	if len(matches) == 0 || err != nil {
		mg.Deps(Build.Deb)
	}

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
