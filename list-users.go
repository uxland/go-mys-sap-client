package mys_sap_client

import (
	"encoding/json"
	"net/http"
)

func (s *sapClient) ListUsers(search string, maxResults int32, auth SAPAuth) ([]SAPUserInfo, error) {
	type body struct {
		Search     string
		MaxResults int32
	}
	args := &fetchArgs{
		url:  "list-users",
		body: &body{MaxResults: maxResults, Search: search},
		user: &auth,
	}
	response, err := s.fetch(args)
	if err != nil {
		return nil, err
	}
	return handleListUsersResponse(response)
}

func handleListUsersResponse(response *http.Response) ([]SAPUserInfo, error) {
	buffer, _, err := assertNoErrorsInResponse(response)
	if err != nil {
		return nil, err
	}
	type modelData struct {
		Data []SAPUserInfo
	}
	model := &modelData{}
	err = json.Unmarshal(buffer, model)
	if err != nil {
		return nil, err
	}
	return model.Data, nil
}
