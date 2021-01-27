package mys_sap_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//Type of authentication
type SAPAuthenticationType = int

const (
	//Basic authentication supplied. user:pwd in base64
	Basic SAPAuthenticationType = iota
	//A SAP cookie is supplied
	Cookies SAPAuthenticationType = 1
)

type SAPCommand struct {
	Name    string
	Payload interface{}
	User    *SAPAuth
}

type SAPUserInfo struct {
	UserID      string `json:"USERNAME"`
	DisplayName string `json:"FULLNAME"`
}

type SAPUser struct {
	UserID  string
	Cookies map[string]string
}

type SAPClient interface {
	SendCommand(cmd *SAPCommand, result interface{}) (*SapResponse, error)
	Authenticate(auth SAPAuth) (*SAPUser, error)
	ListUsers(search string, maxResults int32, auth SAPAuth) ([]SAPUserInfo, error)
}

type sapClient struct {
	URLBase   string
	MySAppID  string
	SAPClient string
}

func NewClient(URLBase string, mysAppID string, SAPClient string) SAPClient {
	return &sapClient{URLBase: URLBase, MySAppID: mysAppID, SAPClient: SAPClient}
}

type fetchArgs struct {
	method      string
	url         string
	body        interface{}
	headers     map[string]string
	queryParams map[string]string
	user        *SAPAuth
}

func (s *sapClient) fetch(args *fetchArgs) (*http.Response, http.CookieJar, error) {
	timeout := 10 * time.Minute
	jar := newSapCookieJar(args.user)

	client := http.Client{
		Timeout: timeout,
		Jar:     jar,
	}
	request, err := s.createRequest(args)
	if err != nil {
		return nil, nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("error fetching SAP:\n%s\n", err.Error())
		return nil, nil, err
	}
	return response, jar, nil
}
func getNoCookie() string {
	const UserCtxCookie = "sap-usercontext"
	const SsoCookie = "MYSAPSSO2"
	const SapSessionCookie = "SAP_SESSIONID_BID_100"
	noCookie := fmt.Sprintf("%s=nop;%s=nop;%s=nop;", UserCtxCookie, SsoCookie, SapSessionCookie)
	return noCookie
}
func (s *sapClient) createRequest(args *fetchArgs) (*http.Request, error) {
	body, err := json.Marshal(args.body)
	if err != nil {
		return nil, err
	}
	URL := s.buildUrl(args)
	request, err := http.NewRequest(args.method, URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("content-type", "application/json")
	if args.user != nil {
		if args.user.Type == Basic {
			request.Header.Set("Authorization", "Basic "+args.user.Value)
		}
	}
	return request, nil
}
func (s *sapClient) buildUrl(args *fetchArgs) string {
	var res = fmt.Sprintf(`%v%v%v`,
		s.URLBase,
		args.url,
		queryParamsBuilder(args))
	return res
}
func queryParamsBuilder(args *fetchArgs) string {
	res := ""
	var cont = 0
	for key, value := range args.queryParams {
		if value != "" {
			if cont == 0 {
				res = "?"
			}
			res = res + fmt.Sprintf(`%v=%v`, key, value)
			if cont < len(args.queryParams)-1 {
				res = res + fmt.Sprintf(`%s`, "&")
			}
		}
		cont++
	}
	return res
}

type SapMessage struct {
	MsgType  string `json:"msgType"`
	MsgTitle string `json:"msgTitle"`
}
type SapResponse struct {
	Result   int    `json:"result"`
	Success  string `json:"SUCCESS"`
	Messages []SapMessage
}
type HttpError struct {
	StatusCode int
	Message    string
}

func (h *HttpError) Error() string {
	return fmt.Sprintf("Http error: %d, %s", h.StatusCode, h.Message)
}
func checkHttpStatusInternal(statusCode int, message string) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	return &HttpError{Message: message, StatusCode: statusCode}
}
func checkHttpStatus(response *http.Response) error {
	return checkHttpStatusInternal(response.StatusCode, response.Status)
}

func checkErrorsInSapResponse(sapResponse *SapResponse) error {
	var err error
	if sapResponse.Result > 0 {
		err = checkHttpStatusInternal(sapResponse.Result, "error")
		if err != nil {
			return err
		}
	}
	message := ""
	for _, s := range sapResponse.Messages {
		if s.MsgType == "E" {
			message = fmt.Sprintf("%s\n%s", message, s.MsgTitle)
		}
	}
	if len(message) > 0 || sapResponse.Success != "X" {
		err = errors.New(message)
	}
	return err
}
func assertNoErrorsInResponse(response *http.Response) ([]byte, *SapResponse, error) {
	err := checkHttpStatus(response)
	if err != nil {
		return nil, nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, nil, err
	}
	m := &SapResponse{}
	err = json.Unmarshal(body, m)
	if err != nil {
		return nil, nil, err
	}
	err = checkErrorsInSapResponse(m)
	if err != nil {
		return nil, m, err
	}
	return body, m, nil
}
