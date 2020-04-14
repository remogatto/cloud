package cloud

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// A client represents a client connection to a {own|next}cloud
type Client struct {
	Url      *url.URL
	Username string
	Password string
}

// Error type encapsulates the returned error messages from the
// server.
type Error struct {
	// Exception contains the type of the exception returned by
	// the server.
	Exception string `xml:"exception"`

	// Message contains the error message string from the server.
	Message string `xml:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("Exception: %s, Message: %s", e.Exception, e.Message)
}

type ShareResult struct {
	XMLName    xml.Name `xml:"ocs"`
	Status     string   `xml:"meta>status"`
	StatusCode uint     `xml:"meta>statuscode"`
	Message    string   `xml:"meta>message"`
	Id         uint     `xml:"data>id"`
}

// Dial connects to an {own|next}Cloud instance at the specified
// address using the given credentials.
func Dial(host, username, password string) (*Client, error) {
	url, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	return &Client{
		Url:      url,
		Username: username,
		Password: password,
	}, nil
}

// Mkdir creates a new directory on the cloud with the specified name.
func (c *Client) Mkdir(path string) error {
	_, err := c.sendWebDavRequest("MKCOL", path, nil)
	return err

}

// Delete removes the specified folder from the cloud.
func (c *Client) Delete(path string) error {
	_, err := c.sendWebDavRequest("DELETE", path, nil)
	return err
}

// Upload uploads the specified source to the specified destination
// path on the cloud.
func (c *Client) Upload(src []byte, dest string) error {
	_, err := c.sendWebDavRequest("PUT", dest, src)
	return err
}

// UploadDir uploads an entire directory on the cloud. It returns the
// path of uploaded files or error. It uses glob pattern in src.
func (c *Client) UploadDir(src string, dest string) ([]string, error) {
	files, err := filepath.Glob(src)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		err = c.Upload(data, filepath.Join(dest, filepath.Base(file)))
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// Download downloads a file from the specified path.
func (c *Client) Download(path string) ([]byte, error) {
	return c.sendWebDavRequest("GET", path, nil)
}

func (c *Client) Exists(path string) bool {
	_, err := c.sendWebDavRequest("PROPFIND", path, nil)
	return err == nil
}

func (c *Client) CreateGroupFolder(mountPoint string) (*ShareResult, error) {
	return c.sendAppsRequest("POST", "groupfolders/folders", fmt.Sprintf("mountpoint=%s", mountPoint))
}

func (c *Client) AddGroupToGroupFolder(group string, folderId uint) (*ShareResult, error) {
	return c.sendAppsRequest("POST", fmt.Sprintf("groupfolders/folders/%d/groups", folderId), fmt.Sprintf("group=%s", group))
}

func (c *Client) SetGroupPermissionsForGroupFolder(permissions int, group string, folderId uint) (*ShareResult, error) {
	return c.sendAppsRequest("POST", fmt.Sprintf("apps/groupfolders/folders/%d/groups/%s", folderId, group), fmt.Sprintf("permissions=%d", permissions))
}

func (c *Client) sendWebDavRequest(request string, path string, data []byte) ([]byte, error) {
	// Create the https request

	webdavPath := filepath.Join("remote.php/webdav", path)

	folderUrl, err := url.Parse(webdavPath)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(request, c.Url.ResolveReference(folderUrl).String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(body) > 0 {
		if body[0] == '<' {
			error := Error{}
			err = xml.Unmarshal(body, &error)
			if err != nil {
				return body, err
			}
			if error.Exception != "" {
				return nil, err
			}
		}

	}

	return body, nil
}

func (c *Client) sendAppsRequest(request string, path string, data string) (*ShareResult, error) {
	// Create the https request

	appsPath := filepath.Join("apps", path)

	folderUrl, err := url.Parse(appsPath)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(request, c.Url.ResolveReference(folderUrl).String(), strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("OCS-APIRequest", "true")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := ShareResult{}
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != 100 {
		return nil, fmt.Errorf("Share API returned an unsuccessful status code %d", result.StatusCode)
	}

	return &result, nil
}
