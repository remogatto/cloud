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
	client, err = Dial(
		"http://localhost:18080/remote.php/webdav/",
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

func (t *testSuite) TestUploadDir() {
	err := client.Mkdir("Test")
	t.Nil(err)

	err = client.Mkdir("Test/Folder")
	t.Nil(err)

	files, err := client.UploadDir(filepath.Join(testDir, "Folder/*"), "Test/Folder/")
	t.Nil(err)

	t.Equal(1, len(files))

	data, err := client.Download("Test/Folder/test.txt")
	t.Nil(err)

	t.Equal("Hello World!\n", string(data))
}

func (t *testSuite) TestExists() {
	err := client.Mkdir("Test")
	t.Nil(err)
	t.True(client.Exists("Test"))
}
