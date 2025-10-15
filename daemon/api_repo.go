package daemon

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"
)

type RepoAPI struct {
	userAgent   string
	baseURL     string
	version     string
	env         internal.Environment
	packageType string
	arch        string
	client      *http.Client
	sync.Mutex
}

type RepoAPIResponse struct {
	Headers http.Header
	Body    io.ReadCloser
}

func NewRepoAPI(
	userAganet string,
	baseURL string,
	version string,
	env internal.Environment,
	packageType,
	arch string,
	client *http.Client,
) *RepoAPI {
	return &RepoAPI{
		userAgent:   userAganet,
		baseURL:     baseURL,
		version:     version,
		env:         env,
		packageType: packageType,
		arch:        arch,
		client:      client,
	}
}

func (api *RepoAPI) DebianFileList() ([]byte, error) {
	repoType := core.RepoTypeProduction

	resp, err := api.request(fmt.Sprintf(core.DebFileinfoURLFormat, repoType, api.arch))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch debian fileinfo: %w", err)
	}
	defer resp.Body.Close()

	body, err := core.MaxBytesReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read debian fileinfo data: %w", err)
	}

	return body, nil
}

func (api *RepoAPI) RpmFileList() ([]byte, error) {
	repoType := core.RepoTypeProduction
	repoArch := "i386"
	switch api.arch {
	case "amd64":
		repoArch = "x86_64"
	case "arm64":
		repoArch = "aarch64"
	case "armel":
		repoArch = "armv5f"
	case "armhf":
		repoArch = "armhfp"
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", api.arch)
	}

	resp, err := api.request(fmt.Sprintf(core.RpmRepoMdURLFormat, repoType, repoArch, core.RpmRepoMdURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repomd: %w", err)
	}
	defer resp.Body.Close()

	body, err := core.MaxBytesReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read repomd data: %w", err)
	}

	filelistPattern := regexp.MustCompile(`/.*filelists\.xml\.gz`)
	filepath := strings.TrimLeft(filelistPattern.FindString(string(body)), "/")

	resp, err = api.request(fmt.Sprintf(core.RpmRepoMdURLFormat, repoType, repoArch, filepath))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rpm fileinfo: %w", err)
	}
	defer resp.Body.Close()

	body, err = core.MaxBytesReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read rpm fileinfo: %w", err)
	}

	return body, nil
}

func (api *RepoAPI) request(path string) (*RepoAPIResponse, error) {
	req, err := request.NewRequest(http.MethodGet, api.userAgent, api.baseURL, path, "", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := core.ExtractError(resp); err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return &RepoAPIResponse{
		Headers: resp.Header,
		Body:    reader,
	}, nil
}
