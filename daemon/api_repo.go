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
	baseURL string,
	version string,
	env internal.Environment,
	packageType,
	arch string,
	client *http.Client,
) *RepoAPI {
	return &RepoAPI{
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
		//log.Printf("DebianFileList failed to fetch fileinfo. Error: %v.\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, core.ReaderLimit))
	if err != nil {
		//log.Printf("DebianFileList failed to read fileinfo data. Error: %v.\n", err)
		return nil, err
	}

	return body, nil
}

func (api *RepoAPI) RpmFileList() ([]byte, error) {
	repoType := core.RepoTypeProduction
	repoArch := "i386"
	if api.arch == "amd64" {
		repoArch = "x86_64"
	} else if api.arch == "arm64" {
		repoArch = "aarch64"
	} else if api.arch == "armel" {
		repoArch = "armv5f"
	} else if api.arch == "armhf" {
		repoArch = "armhfp"
	} else {
		return nil, fmt.Errorf("unsupported architecture: %s", api.arch)
	}

	resp, err := api.request(fmt.Sprintf(core.RpmRepoMdURLFormat, repoType, repoArch, core.RpmRepoMdURL))
	if err != nil {
		//log.Printf("RpmFileList failed to fetch repomd. Error: %v.\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, core.ReaderLimit))
	if err != nil {
		//log.Printf("RpmFileList failed to read repomd data. Error: %v.\n", err)
		return nil, err
	}

	filelistPattern := regexp.MustCompile(`/.*filelists\.xml\.gz`)
	filepath := strings.TrimLeft(filelistPattern.FindString(string(body)), "/")

	resp, err = api.request(fmt.Sprintf(core.RpmRepoMdURLFormat, repoType, repoArch, filepath))
	if err != nil {
		//log.Printf("RpmFileList failed to fetch fileinfo. Error: %v.\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(io.LimitReader(resp.Body, core.ReaderLimit))
	if err != nil {
		//log.Printf("RpmFileList failed to read fileinfo. Error: %v.\n", err)
		return nil, err
	}

	return body, nil
}

func (api *RepoAPI) request(path string) (*RepoAPIResponse, error) {
	req, err := request.NewRequest(http.MethodGet, fmt.Sprintf("NordApp Linux %s %s", api.version, api.packageType), api.baseURL, path, "", "", "gzip, deflate", nil)
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
