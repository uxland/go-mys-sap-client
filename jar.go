package mys_sap_client

import (
	"net/http"
	"net/url"
)

type sapCookieJar struct {
	cookies []*http.Cookie
}

func newSapCookieJar(auth *SAPAuth) http.CookieJar {
	var cookies []*http.Cookie
	if auth != nil && auth.Type == Cookies {
		//strings.Split(auth.Value, ";")
		for s, cookieValue := range auth.Cookies {
			cookies = append(cookies, &http.Cookie{Name: s, Value: cookieValue})
		}

	}
	return &sapCookieJar{cookies}
}

func (s *sapCookieJar) SetCookies(_ *url.URL, cookies []*http.Cookie) {
	//panic("implement me")
	s.cookies = cookies
}

func (s *sapCookieJar) Cookies(_ *url.URL) []*http.Cookie {
	return s.cookies
}

func toCookiesMap(cookies []*http.Cookie) map[string]string {
	result := make(map[string]string)
	for _, cookie := range cookies {
		result[cookie.Name] = cookie.Value
	}
	return result
}
