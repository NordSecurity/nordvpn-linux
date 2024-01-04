// Package pulp provides package repository management functionality.
package pulp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	apiPrefix  = "/pulp/api/v2"
	loginURL   = apiPrefix + "/actions/login/"
	packageURL = apiPrefix + "/repositories/%s/search/units/"
	removeURL  = apiPrefix + "/repositories/%s/actions/unassociate/"

	// filter has name in order to exclude nordvpn-release package
	filter = `{"criteria": {"type_ids": [%q], "filters": {"unit": {"name": "nordvpn"}}}}`
	// picker has both name and version in order to return only a single package
	picker = `{"criteria": {"type_ids": [%q], "filters": {"unit": {"$and": [{"version": %q}, {"name": "nordvpn"}]}}}}`
)

type loginResponse struct {
	// PublicCert is a public part of a certificate.
	PublicCert string `json:"certificate"`
	PrivateKey string `json:"key"`
}

func (l *loginResponse) certificate() string {
	// https://github.com/pulp/pulp/blob/2-master/client_admin/pulp/client/admin/admin_auth.py#L50
	return l.PrivateKey + l.PublicCert
}

type packageResponse struct {
	Metadata struct {
		Version string `json:"version"`
	} `json:"metadata"`
}

// Login return client-side certificate authorized http client.
func Login(
	hostname string,
	username string,
	password string,
	caDER []byte,
) (*http.Client, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("cert pool: %w", err)
	}

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, fmt.Errorf("parsing der: %w", err)
	}
	pool.AddCert(caCert)

	req, err := http.NewRequest(http.MethodPost, hostname+loginURL, bytes.NewReader(nil))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	// transport is an interface, which means that updating tls config requires
	// casting, so we don't update it and create http client for a single request
	// instead
	resp, err := (&http.Client{
		Transport: &http.Transport{
			// #nosec G402 -- minimum tls version is controlled by the standard library
			TLSClientConfig: &tls.Config{
				Renegotiation: tls.RenegotiateFreelyAsClient,
				RootCAs:       pool,
			},
		},
	}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var parsed loginResponse
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		log.Println(string(data))
		return nil, fmt.Errorf("decode: %w", err)
	}

	cert, err := tls.X509KeyPair([]byte(parsed.certificate()), []byte(parsed.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("keypair: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			// #nosec G402 -- minimum tls version is controlled by the standard library
			TLSClientConfig: &tls.Config{
				Certificates:  []tls.Certificate{cert},
				Renegotiation: tls.RenegotiateFreelyAsClient,
				RootCAs:       pool,
			},
		},
	}, nil
}

// Debs which can be deleted.
func Debs(
	client *http.Client,
	hostname string,
	repository string,
	count uint,
) ([]string, error) {
	if strings.Contains(repository, "centos") {
		return nil, fmt.Errorf("%s does not support deb packages", repository)
	}
	return versions(client, hostname, repository, fmt.Sprintf(filter, "deb"), count)
}

// Rpms which can be deleted.
func Rpms(
	client *http.Client,
	hostname string,
	repository string,
	count uint,
) ([]string, error) {
	if strings.Contains(repository, "debian") {
		return nil, fmt.Errorf("%s does not support rpm packages", repository)
	}
	return versions(client, hostname, repository, fmt.Sprintf(filter, "rpm"), count)
}

func RemoveDeb(
	client *http.Client,
	hostname string,
	repository string,
	version string,
) error {
	if strings.Contains(repository, "centos") {
		return fmt.Errorf("%s does not support deb packages", repository)
	}
	return remove(client, hostname, repository, fmt.Sprintf(picker, "deb", version))
}

func RemoveRpm(
	client *http.Client,
	hostname string,
	repository string,
	version string,
) error {
	if strings.Contains(repository, "debian") {
		return fmt.Errorf("%s does not support rpm packages", repository)
	}
	return remove(client, hostname, repository, fmt.Sprintf(picker, "rpm", version))
}

func remove(client *http.Client, hostname string, repository string, criteria string) error {
	if repository == "" {
		return errors.New("repository not provided")
	}

	resp, err := client.Post(
		fmt.Sprintf(hostname+removeURL, repository),
		"application/json",
		bytes.NewReader([]byte(criteria)),
	)

	if err != nil {
		return fmt.Errorf("post: %w", err)
	}

	defer resp.Body.Close()
	if !(resp.StatusCode < 400) {
		return errors.New(http.StatusText(resp.StatusCode))
	}
	return nil
}

func versions(
	client *http.Client,
	hostname string,
	repository string,
	criteria string,
	count uint,
) ([]string, error) {
	if repository == "" {
		return nil, errors.New("repository not provided")
	}

	resp, err := client.Post(
		fmt.Sprintf(hostname+packageURL, repository),
		"application/json",
		bytes.NewReader([]byte(criteria)),
	)

	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	defer resp.Body.Close()

	var data []packageResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(string(body))
		return nil, fmt.Errorf("decode: %w", err)
	}

	var ret []string
	for _, elem := range data {
		ret = append(ret, elem.Metadata.Version)
	}

	return deleteFrom(ret, count), err
}
