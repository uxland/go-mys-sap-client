package mys_sap_client

import (
	"encoding/base64"
	"testing"
)

const url = "http://35.232.188.122/qua/api/"
const sapClientID = "100"

func TestAuthenticate(t *testing.T) {
	client := NewClient(url, "", sapClientID)
	auth := "uxland1:uxland2020!"
	basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	sapUser, err := client.Authenticate(SAPAuth{Type: Basic, Value: basicAuth})
	if err != nil {
		t.Error(err)
	}
	if sapUser == nil {
		t.Error("user not found")
	}
}
func TestAuthenticateError(t *testing.T) {
	client := NewClient(url, "", sapClientID)
	auth := "uxland1:uxland2019!"
	basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	sapUser, err := client.Authenticate(SAPAuth{Type: Basic, Value: basicAuth})
	if err == nil {
		t.Error(err)
	}
	if sapUser != nil {
		t.Error("user found")
	}
}

func TestAuthenticateByCookies(t *testing.T) {
	client := NewClient(url, "", sapClientID)
	auth := "uxland1:uxland2020!"
	basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	sapUser, _ := client.Authenticate(SAPAuth{Type: Basic, Value: basicAuth})
	cookie := buildSapCookies(sapUser.Cookies)
	sapUser2, err := client.Authenticate(SAPAuth{Type: Cookies, Value: cookie})
	if err != nil {
		t.Error(err)
	}
	if sapUser2 == nil {
		t.Error("user not found")
	}
	if sapUser2.UserID != sapUser.UserID {
		t.Error()
	}
	cookie = buildSapCookies(sapUser2.Cookies)
	sapUser3, err := client.Authenticate(SAPAuth{Type: Cookies, Value: cookie})
	if err != nil {
		t.Error(err)
	}
	if sapUser3 == nil {
		t.Error("user not found")
	}
	if sapUser3.UserID != sapUser.UserID {
		t.Error()
	}
}
