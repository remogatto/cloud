package cloud

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/remogatto/prettytest"
)

type testSuite struct {
	prettytest.Suite
}

var (
	client  *Client
	testDir = "test/testdata"
)

func TestRunner(t *testing.T) {
	prettytest.Run(
		t,
		new(testSuite),
	)
}

func (t *testSuite) BeforeAll() {
	var err error
	client, err = NewClient(
		"http://localhost:8080/remote.php/webdav/",
		"admin",
		"password",
	)
	if err != nil {
		panic(err)
	}

}

func (t *testSuite) After() {
	client.Delete("Test")
}

func (t *testSuite) TestMkDir() {
	err := client.Mkdir("Test")
	t.Nil(err)
}

func (t *testSuite) TestDelete() {
	err := client.Mkdir("Test")
	t.Nil(err)
	err = client.Delete("Test")
	t.Nil(err)
}

func (t *testSuite) TestDownloadUpload() {
	err := client.Mkdir("Test")
	t.Nil(err)

	src, err := ioutil.ReadFile(filepath.Join(testDir, "test.txt"))
	err = client.Upload(src, "Test/test.txt")
	t.Nil(err)

	data, err := client.Download("Test/test.txt")
	t.Nil(err)

	t.Equal("Hello World!\n", string(data))
}
