package sfs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"starlight/common/httpclient"
	"strconv"
	"strings"
	"testing"
)

const (
	testLoginAPI   string = "http://uat.starlight-dev.nscc-gz.cn/api/keystone/short_term_token/name"
	testUsername   string = "nscc-gz_yfb_2"
	testPassword   string = "UHcyMDIyUkQ="
	testBaseDir    string = "/HOME/nscc-gz_yfb_2/testDir1"
	testOtherFSDir string = "/GPUFS/nscc-gz_yfb_2/testDir1"
	testFile       string = "/home/zetako/git/cwl.go/intergration/sfs/sfs.go"
)

var (
	globalSFS    *StarlightFileSystem
	fileContent  []byte
	fileCheckSum string
)

func getToken() (string, error) {
	jsonBody := fmt.Sprintf("{\"username\":\"%v\",\"password\":\"%v\"}", testUsername, testPassword)
	resp, err := http.Post(testLoginAPI, "application/json;charset=UTF-8", strings.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http code not 200, but %d", resp.StatusCode)
	}
	var token string
	_, err = httpclient.GetSpecResponse(resp.Body, &token)
	if err != nil {
		return "", err
	}
	//log.Println(token)
	return token, nil
	//return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MjE2NzIsImlwIjoiMTcyLjE2LjE3MS4zNyIsIm9yaV91c2VyX25hbWUiOiIiLCJ1c2VyX25hbWUiOiJuc2NjLWd6X3lmYl8yIiwidWlkIjoxMDAxMSwiZ3JvdXBfbmFtZSI6Im5zY2MtZ3pfeWZiIiwiZ3JvdXBfaWQiOjk4NjUsInN0YXR1cyI6MTEsImV4cCI6MTY4NzkyOTEzNX0.n2K_zsG0G5GL67WhwtfnyxzyZ9uk0i70FmgsCL2GvFZn37DZtQICPRpGIjT_AxQG7u9l71vlQvsqYzlCp4zz4Q", nil
}

func init() {
	// 1. get token
	token, err := getToken()
	if err != nil {
		panic(err)
	}
	// 2. get sfs
	globalSFS, err = NewStarlightFileSystem(context.TODO(), token)
	if err != nil {
		panic(err)
	}
	// 3. test dirs should exist and empty
	shouldEmpty, err := globalSFS.DirInfo(testBaseDir, 1)
	if err != nil {
		panic(err)
	}
	if len(shouldEmpty.Listing) > 0 {
		panic(fmt.Errorf("test dir 1 is not EMPTY"))
	}
	shouldEmpty, err = globalSFS.DirInfo(testOtherFSDir, 1)
	if err != nil {
		panic(err)
	}
	if len(shouldEmpty.Listing) > 0 {
		panic(fmt.Errorf("test dir 2 is not EMPTY"))
	}
	// 4. read file content
	f, err := os.Open(testFile)
	if err != nil {
		panic(err)
	}
	fileContent, err = io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
	// 5. file checksum
}

func Test_SFS_Create(t *testing.T) {
	// 1. Normal upload
	file, err := globalSFS.Create(path.Join(testBaseDir, "create_test.file"), string(fileContent))
	if err != nil {
		t.Fatalf("Upload Failed: %v", err)
	}
	if file.Checksum != fileCheckSum {
		t.Fatalf("CheckSum not Matched")
	}
	// 2. Overwrite with a shorter one
	// TODO no overwrite permitted now
}
func Test_SFS_Info(t *testing.T) {
	// 0. Generate file
	testPath := path.Join(testBaseDir, "info_test.file")
	testNotExist := path.Join(testBaseDir, "never_exist.file")
	file, err := globalSFS.Create(testPath, string(fileContent))
	if err != nil {
		t.Fatalf("Upload Failed: %v", err)
	}
	if file.Checksum != fileCheckSum {
		t.Fatalf("CheckSum not Matched")
	}
	// 1. Get file info just uploaded
	info, err := globalSFS.Info(testPath)
	if err != nil {
		t.Fatalf("Get Info Error: %v", err)
	}
	t.Logf("%+v", info)
	// 2. Get file info never exist
	info, err = globalSFS.Info(testNotExist)
	if err == nil {
		t.Fatalf("Expect Got Info but not")
	}
	t.Logf("Got Wanted Error:\n%+v", err)
}
func Test_SFS_EnsureDir(t *testing.T) {
	// 1. Ensure a dir
	err := globalSFS.EnsureDir(path.Join(testBaseDir, "ensure_test.dir"), os.ModePerm)
	if err != nil {
		t.Fatalf("Ensure Dir Error: %v", err)
	}
	// 2. Ensure an already exist dir
	err = globalSFS.EnsureDir(path.Join(testBaseDir, "ensure_test.dir"), os.ModePerm)
	if err != nil {
		t.Fatalf("Ensure Dir Error: %v", err)
	}
}
func Test_SFS_DirInfo(t *testing.T) {
	// 0. Construct a 5-tier dir tree struct
	lastTierDir := testBaseDir
	for i := 1; i <= 5; i++ {
		lastTierDir = path.Join(lastTierDir, "tier"+strconv.Itoa(i))
		err := globalSFS.EnsureDir(lastTierDir, os.ModePerm)
		if err != nil {
			t.Fatalf("Ensure Dir Error: %v", err)
		}
		_, err = globalSFS.Create(path.Join(lastTierDir, "dir_info_test.file"), string(fileContent))
		if err != nil {
			t.Fatalf("Upload Failed: %v", err)
		}
	}
	// 1. Tree it
	dirInfo, err := globalSFS.DirInfo(testBaseDir, 10)
	if err != nil {
		t.Fatalf("Get DirInfo Failed: %v", err)
	}
	t.Log(dirInfo) // TODO Need to check result
	// 2. Tree it with 3 deep
	dirInfo, err = globalSFS.DirInfo(testBaseDir, 3)
	if err != nil {
		t.Fatalf("Get DirInfo Failed: %v", err)
	}
	t.Log(dirInfo) // TODO Need to check result
}
func Test_SFS_Copy(t *testing.T) {
	// 0. Generate file
	srcFile := path.Join(testBaseDir, "copy_test.file")
	_, err := globalSFS.Create(srcFile, string(fileContent))
	if err != nil {
		t.Fatalf("Upload Failed: %v", err)
	}
	// 1. Copy within FS
	err = globalSFS.Copy(srcFile, path.Join(testBaseDir, "copy_test.file.duplicate"))
	if err != nil {
		t.Fatalf("Copy Failed: %v", err)
	}
	// 2. Copy across FS
	err = globalSFS.Copy(srcFile, path.Join(testOtherFSDir, "copy_test.file.duplicate"))
	if err != nil {
		t.Fatalf("Copy Failed: %v", err)
	}
}
func Test_SFS_Contents(t *testing.T) {
	// 0. Generate file
	// 1. Get uploaded file
}
func Test_SFS_Glob(t *testing.T) {
	// 0. Generate 5 files
	fileNames := []string{
		"glob_test.file.123",    // matched
		"glob_test.file.1231",   // matched
		"glob_test.file.12312",  // matched
		"glob_test.file.1212",   // NOT matched
		"glob_test.file.121312", // NOT matched
	}
	for _, name := range fileNames {
		_, err := globalSFS.Create(path.Join(testBaseDir, name), string(fileContent))
		if err != nil {
			t.Fatalf("Upload Failed: %v", err)
		}
	}
	// 1. Use pattern to get 3 of them
	// TODO Glob has workdir problem
}
func Test_SFS_Migrate(t *testing.T) {
	// 0. Generate file
	srcFile := path.Join(testBaseDir, "copy_test.file")
	_, err := globalSFS.Create(srcFile, string(fileContent))
	if err != nil {
		t.Fatalf("Upload Failed: %v", err)
	}
	// 1. Migrate within FS
	err = globalSFS.Migrate(srcFile, path.Join(testBaseDir, "copy_test.file.duplicate"))
	if err != nil {
		t.Fatalf("Copy Failed: %v", err)
	}
	// 2. Migrate across FS
	err = globalSFS.Migrate(srcFile, path.Join(testOtherFSDir, "copy_test.file.duplicate"))
	if err != nil {
		t.Fatalf("Copy Failed: %v", err)
	}
}
