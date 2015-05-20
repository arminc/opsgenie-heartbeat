package opsgenie

import (
	"testing"
	"time"
)

var testargs = OpsArgs{"testKey", "testName", "testDescription", 99, "month", time.Second * 10, true}

func TestCreateUrl(t *testing.T) {
	var requestParams = make(map[string]string)
	requestParams["apiKey"] = "test"
	url, err := createURL("/v1/test", requestParams)
	if err != nil {
		t.Errorf(err.Error())
	}
	testURL := "https://api.opsgenie.com/v1/test?apiKey=test"
	if url != testURL {
		t.Errorf("Url not correct is [%s] but should be [%s]", url, testURL)
	}
}

func TestAllContentParams(t *testing.T) {
	var all = allContentParams(testargs)
	if all["apiKey"] != testargs.ApiKey || all["name"] != testargs.Name || all["description"] != testargs.Description || all["interval"] != testargs.Interval || all["intervalUnit"] != testargs.IntervalUnit {
		t.Errorf("OpsArgs [%+v] are not the same as all content params [%s]", testargs, all)
	}
}

func TestMandatoryRequestParams(t *testing.T) {
	var params = mandatoryRequestParams(testargs)
	if params["apiKey"] != testargs.ApiKey || params["name"] != testargs.Name {
		t.Errorf("Requested params [%s] are not the same as from OpsArgs [%+v]", params, testargs)
	}
}

func TestCreateErrorResponse(t *testing.T) {
	json := `{"code":10, "error": "test error"}`
	errorResp, err := createErrorResponse([]byte(json))
	if err != nil {
		t.Errorf(err.Error())
	}
	if errorResp.Code != 10 || errorResp.Message != "test error" {
		t.Errorf("Error [%+v] does not correspond to json [%s]", errorResp, json)
	}
}
