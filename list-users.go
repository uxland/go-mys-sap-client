package mys_sap_client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (s *sapClient) ListUsers(search string, maxResults int32, auth SAPAuth) ([]SAPUserInfo, error) {
	args := &fetchArgs{
		url: "user-list",
		queryParams: map[string]string{
			"SEARCH":     search,
			"MAXRESULTS": strconv.Itoa(int(maxResults)),
			"sap-client": s.SAPClient,
		},
		user: &auth,
	}
	response, _, err := s.fetch(args)
	if err != nil {
		return nil, err
	}
	return handleListUsersResponse(response)
}

func handleListUsersResponse(response *http.Response) ([]SAPUserInfo, error) {
	err := checkHttpStatus(response)
	if err != nil {
		return nil, err
	}
	buffer, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	var model []SAPUserInfo
	err = json.Unmarshal(buffer, &model)
	if err != nil {
		return nil, err
	}
	return model, nil
}
