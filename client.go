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

type SAPClient interface {
	SendCommand(cmd *SAPCommand, result interface{}) error
}

type sapClient struct {
	URLBase   string
	MySAppID  string
	SAPClient string
}

func NewClient(URLBase string, mysAppID string, SAPClient string) SAPClient {
	return &sapClient{URLBase: URLBase, MySAppID: mysAppID, SAPClient: SAPClient}
}

func (s *sapClient) SendCommand(cmd *SAPCommand, result interface{}) error {
	args := &fetchArgs{
		method:  "POST",
		url:     "command",
		body:    cmd.Payload,
		headers: nil,
		queryParams: map[string]string{
			"APPID":      s.MySAppID,
			"COMMAND":    cmd.Name,
			"sap-client": s.SAPClient,
		},
		user: cmd.User,
	}
	return s.fetch(args, result)
}

type fetchArgs struct {
	method      string
	url         string
	body        interface{}
	headers     map[string]string
	queryParams map[string]string
	user        *SAPAuth
}

func (s *sapClient) fetch(args *fetchArgs, result interface{}) error {
	timeout := 10 * time.Minute
	client := http.Client{
		Timeout: timeout,
	}
	request, err := s.createRequest(args)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("error fetching SAP:\n%s\n", err.Error())
		return err
	}
	return handlerResponse(response, result)
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
		} else {
			request.Header.Set("cookie", args.user.Value)
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

type sapMessage struct {
	MsgType  string `json:"msgType"`
	MsgTitle string `json:"msgTitle"`
}
type sapModel struct {
	Data     string `json:"DATA"`
	Result   int    `json:"result"`
	Success  string `json:"SUCCESS"`
	Messages []sapMessage
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

func checkErrorsInSapResponse(sapResponse *sapModel) error {
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
	if len(message) > 0 {
		err = errors.New(message)
	}
	return err
}
func handlerResponse(response *http.Response, result interface{}) error {
	err := checkHttpStatus(response)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return err
	}
	m := &sapModel{}
	err = json.Unmarshal(body, m)
	if err != nil {
		return err
	}
	err = checkErrorsInSapResponse(m)
	if err != nil {
		return err
	}
	data := []byte(m.Data)
	return json.Unmarshal(data, result)
}
