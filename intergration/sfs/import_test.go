package sfs

import "testing"

var (
	testAppServiceApi string = "https://uat.starlight-dev.nscc-gz.cn/api/app/cwl/"
	testPrivateApp    string = "未发布的cat"
	testPublicApp     string = "已发布的cat"
)

func Test_Import(t *testing.T) {
	_, err := globalSFS.Load(testAppServiceApi + testPublicApp)
	if err != nil {
		t.Fatalf("Load CWL Doc failed: %v", err)
	}
	_, err = globalSFS.Load(testAppServiceApi + testPrivateApp)
	if err != nil {
		t.Fatalf("Load CWL Doc failed: %v", err)
	}
}
