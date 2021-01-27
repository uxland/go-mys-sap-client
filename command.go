package mys_sap_client

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (s *sapClient) SendCommand(cmd *SAPCommand, result interface{}) (*SapResponse, error) {
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
	response, _, err := s.fetch(args)
	if err != nil {
		return nil, err
	}
	return handleCommand(response, result)
}

func handleCommand(response *http.Response, result interface{}) (*SapResponse, error) {
	buffer, sapResponse, err := assertNoErrorsInResponse(response)
	if err != nil {
		return sapResponse, err
	}
	type commandModel struct {
		*SapResponse
		Data interface{} `json:"DATA,omitempty"`
	}
	model := &commandModel{Data: result}
	err = json.Unmarshal(buffer, model)
	if err != nil {
		var unmarshallError *json.UnmarshalTypeError
		if errors.As(err, &unmarshallError) {
			if unmarshallError.Value == "string" {
				err = nil
			}
		}
	}
	return model.SapResponse, err
}
