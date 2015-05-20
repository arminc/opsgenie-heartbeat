package opsgenie

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
)

var timeout = time.Second * 30
var apiURL = "https://api.opsgenie.com"

func startHeartbeatAndSend(args OpsArgs) {
	startHeartbeat(args)
	sendHeartbeat(args)
}

func startHeartbeat(args OpsArgs) {
	heartbeat, err := getHeartbeat(args)
	if err != nil {
		log.Error(err)
	} else {
		if heartbeat == nil {
			addHeartbeat(args)
		} else {
			updateHeartbeatWithEnabledTrue(args, *heartbeat)
		}
	}
}

//StartHeartbeatLoop can be used from other codes as a library call
func StartHeartbeatLoop(args OpsArgs) {
	startHeartbeat(args)
	sendHeartbeatLoop(args)
}

func getHeartbeat(args OpsArgs) (*Heartbeat, error) {
	code, body, err := doHTTPRequest("GET", "/v1/json/heartbeat/", mandatoryRequestParams(args), nil)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return checkHeartbeatError(code, body, args.name)
	}
	return createHeartbeat(body, args.name)
}

func checkHeartbeatError(code int, body []byte, name string) (*Heartbeat, error) {
	errorResponse, err := createErrorResponse(body)
	if err != nil {
		return nil, err
	}
	if code == 400 && errorResponse.Code == 17 {
		log.Infof("Heartbeat [%s] doesn't exist", name)
		return nil, nil
	}
	return nil, fmt.Errorf("%#v", errorResponse)
}

func createHeartbeat(body []byte, name string) (*Heartbeat, error) {
	heartbeat := &Heartbeat{}
	err := json.Unmarshal(body, &heartbeat)
	if err != nil {
		return nil, err
	}
	log.Info("Successfully retrieved heartbeat [" + name + "]")
	return heartbeat, nil
}

func addHeartbeat(args OpsArgs) {
	doOpsGenieHTTPRequestHandled("POST", "/v1/json/heartbeat/", nil, allContentParams(args), "Successfully added heartbeat ["+args.name+"]")
}

func updateHeartbeatWithEnabledTrue(args OpsArgs, heartbeat Heartbeat) {
	var contentParams = allContentParams(args)
	contentParams["id"] = heartbeat.ID
	contentParams["enabled"] = true
	doOpsGenieHTTPRequestHandled("POST", "/v1/json/heartbeat", nil, contentParams, "Successfully enabled and updated heartbeat ["+args.name+"]")
}

func sendHeartbeat(args OpsArgs) {
	doOpsGenieHTTPRequestHandled("POST", "/v1/json/heartbeat/send", nil, mandatoryContentParams(args), "Successfully sent heartbeat ["+args.name+"]")
}

func sendHeartbeatLoop(args OpsArgs) {
	for _ = range time.Tick(args.loopInterval) {
		sendHeartbeat(args)
	}
}

func stopHeartbeat(args OpsArgs) {
	if args.delete {
		deleteHeartbeat(args)
	} else {
		disableHeartbeat(args)
	}
}

func deleteHeartbeat(args OpsArgs) {
	doOpsGenieHTTPRequestHandled("DELETE", "/v1/json/heartbeat", mandatoryRequestParams(args), nil, "Successfully deleted heartbeat ["+args.name+"]")
}

func disableHeartbeat(args OpsArgs) {
	doOpsGenieHTTPRequestHandled("POST", "/v1/json/heartbeat/disable", nil, mandatoryContentParams(args), "Successfully disabled heartbeat ["+args.name+"]")
}

func mandatoryContentParams(args OpsArgs) map[string]interface{} {
	var contentParams = make(map[string]interface{})
	contentParams["apiKey"] = args.apiKey
	contentParams["name"] = args.name
	return contentParams
}

func allContentParams(args OpsArgs) map[string]interface{} {
	var contentParams = mandatoryContentParams(args)
	if args.description != "" {
		contentParams["description"] = args.description
	}
	if args.interval != 0 {
		contentParams["interval"] = args.interval
	}
	if args.intervalUnit != "" {
		contentParams["intervalUnit"] = args.intervalUnit
	}
	return contentParams
}

func mandatoryRequestParams(args OpsArgs) map[string]string {
	var requestParams = make(map[string]string)
	requestParams["apiKey"] = args.apiKey
	requestParams["name"] = args.name
	return requestParams
}

func createErrorResponse(responseBody []byte) (ErrorResponse, error) {
	errResponse := &ErrorResponse{}
	err := json.Unmarshal(responseBody, &errResponse)
	if err != nil {
		return *errResponse, err
	}
	return *errResponse, nil
}

func doOpsGenieHTTPRequestHandled(method string, urlSuffix string, requestParameters map[string]string, contentParameters map[string]interface{}, msg string) {
	_, err := doOpsGenieHTTPRequest(method, urlSuffix, requestParameters, contentParameters)
	if err != nil {
		log.Error(err)
	} else {
		log.Info(msg)
	}
}

func doOpsGenieHTTPRequest(method string, urlSuffix string, requestParameters map[string]string, contentParameters map[string]interface{}) ([]byte, error) {
	code, body, err := doHTTPRequest(method, urlSuffix, requestParameters, contentParameters)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		e, err := createErrorResponse(body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%#v", e)
	}
	return body, nil
}

func doHTTPRequest(method string, urlSuffix string, requestParameters map[string]string, contentParameters map[string]interface{}) (int, []byte, error) {
	request, err := createRequest(method, urlSuffix, requestParameters, contentParameters)
	if err != nil {
		return 0, nil, err
	}
	resp, err := getHTTPClient().Do(request)
	if err != nil {
		return 0, nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, body, nil
}

func createRequest(method string, urlSuffix string, requestParameters map[string]string, contentParameters map[string]interface{}) (*http.Request, error) {
	body, err := json.Marshal(contentParameters)
	if err != nil {
		return nil, err
	}
	url, err := createURL(urlSuffix, requestParameters)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return request, nil
}

func createURL(urlSuffix string, requestParameters map[string]string) (string, error) {
	var URL *url.URL
	URL, err := url.Parse(apiURL + urlSuffix)
	if err != nil {
		return "", err
	}
	parameters := url.Values{}
	for k, v := range requestParameters {
		parameters.Add(k, v)
	}
	URL.RawQuery = parameters.Encode()
	return URL.String(), nil
}

func getHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, timeout)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(timeout))
				return conn, nil
			},
		},
	}
	return client
}

//Heartbeat represents the OpsGenie heartbeat data structure
type Heartbeat struct {
	ID string `json:"id"`
}

//ErrorResponse represents the OpsGenie error response data structure
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}
