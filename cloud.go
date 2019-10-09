package cloud

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
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
	_, err := c.sendRequest("MKCOL", path, nil)
	return err

}

// Delete removes the specified folder from the cloud.
func (c *Client) Delete(path string) error {
	_, err := c.sendRequest("DELETE", path, nil)
	return err
}

// Upload uploads the specified source to the specified destination
// path on the cloud.
func (c *Client) Upload(src []byte, dest string) error {
	_, err := c.sendRequest("PUT", dest, src)
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

	pathUrl, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Create the https request

	client := &http.Client{}
	req, err := http.NewRequest("GET", c.Url.ResolveReference(pathUrl).String(), nil)
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

	error := Error{}
	err = xml.Unmarshal(body, &error)
	if err == nil {
		if error.Exception != "" {
			return nil, err
		}
	}

	return body, nil
}

func (c *Client) Exists(path string) bool {
	_, err := c.sendRequest("PROPFIND", path, nil)
	return err == nil
}

func (c *Client) sendRequest(request string, path string, data []byte) ([]byte, error) {
	// Create the https request

	folderUrl, err := url.Parse(path)
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
		error := Error{}
		err = xml.Unmarshal(body, &error)
		if err != nil {
			return body, err
		}
		if error.Exception != "" {
			return nil, err
		}

	}

	return body, nil
}
