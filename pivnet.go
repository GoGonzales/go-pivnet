package pivnet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/pivotal-golang/lager"
)

const (
	Endpoint = "https://network.pivotal.io"
	path     = "/api/v2"
)

type Client interface {
	ProductVersions(productSlug string, releases []Release) ([]string, error)
	CreateRelease(CreateReleaseConfig) (Release, error)
	ReleasesForProductSlug(string) ([]Release, error)
	GetRelease(string, string) (Release, error)
	UpdateRelease(string, Release) (Release, error)
	DeleteRelease(Release, string) error
	GetProductFiles(Release) (ProductFiles, error)
	GetProductFile(productSlug string, releaseID int, productID int) (ProductFile, error)
	EULAs() ([]EULA, error)
	AcceptEULA(productSlug string, releaseID int) error
	CreateProductFile(CreateProductFileConfig) (ProductFile, error)
	DeleteProductFile(productSlug string, id int) (ProductFile, error)
	AddProductFile(productID int, releaseID int, productFileID int) error
	FindProductForSlug(slug string) (Product, error)
	UserGroups(productSlug string, releaseID int) ([]UserGroup, error)
	AddUserGroup(productSlug string, releaseID int, userGroupID int) error
	ReleaseETag(string, Release) (string, error)
	ReleaseTypes() ([]string, error)
}

type client struct {
	url       string
	token     string
	userAgent string
	logger    lager.Logger
}

type NewClientConfig struct {
	Endpoint  string
	Token     string
	UserAgent string
}

func NewClient(config NewClientConfig, logger lager.Logger) Client {
	url := fmt.Sprintf("%s%s", config.Endpoint, path)

	return &client{
		url:       url,
		token:     config.Token,
		userAgent: config.UserAgent,
		logger:    logger,
	}
}

func (c client) ProductVersions(productSlug string, releases []Release) ([]string, error) {
	var versions []string
	for _, r := range releases {
		etag, err := c.ReleaseETag(productSlug, r)
		if err != nil {
			return nil, err
		}
		version := fmt.Sprintf("%s#%s", r.Version, etag)
		versions = append(versions, version)
	}

	return versions, nil
}

func (c client) makeRequestWithHTTPResponse(
	requestType string,
	url string,
	expectedStatusCode int,
	body io.Reader,
	data interface{},
) (*http.Response, error) {
	req, err := http.NewRequest(requestType, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
	req.Header.Add("User-Agent", c.userAgent)

	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		c.logger.Debug("Error dumping request", lager.Data{"error": err})
		return nil, err
	}

	c.logger.Debug("Making request", lager.Data{"request": string(reqBytes)})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Debug("Error making request", lager.Data{"error": err})
		return nil, err
	}
	defer resp.Body.Close()

	c.logger.Debug("Response status code", lager.Data{"status code": resp.StatusCode})
	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf(
			"Pivnet returned status code: %d for the request - expected %d",
			resp.StatusCode,
			expectedStatusCode,
		)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(b) > 0 {
		c.logger.Debug("Response body", lager.Data{"response body": string(b)})
		err = json.Unmarshal(b, data)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (c client) makeRequest(
	requestType string,
	url string,
	expectedStatusCode int,
	body io.Reader,
	data interface{},
) error {
	_, err := c.makeRequestWithHTTPResponse(
		requestType,
		url,
		expectedStatusCode,
		body,
		data,
	)
	return err
}
