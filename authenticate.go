package mys_sap_client

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *sapClient) Authenticate(auth SAPAuth) (*SAPUser, error) {
	args := &fetchArgs{url: "user-info", user: &auth}
	type response struct {
	}
	//res := &response{}
	resp, err := s.fetch(args)
	if err != nil {
		return nil, err
	}
	return handleAuthentication(resp)
}

func extractCookies(response *http.Response) map[string]string {
	header := response.Header.Get("set-cookie")
	cookies := strings.Split(header, ",")
	cookiesMap := make(map[string]string)
	for _, s := range cookies {
		cookie := strings.Split(strings.TrimSpace(s), "=")
		key := cookie[0]
		value := strings.Join(cookie[1:], "=")
		cookiesMap[key] = value

	}
	return cookiesMap
}

func handleAuthentication(response *http.Response) (*SAPUser, error) {
	buffer, _, err := assertNoErrorsInResponse(response)
	if err != nil {
		return nil, err
	}
	type authModel struct {
		UserData struct {
			UserID string `json:"USERNAME"`
		} `json:"USER_DATA"`
	}
	model := &authModel{}
	err = json.Unmarshal(buffer, model)
	if err != nil {
		return nil, err
	}
	return &SAPUser{UserID: model.UserData.UserID, Cookies: extractCookies(response)}, nil
}
