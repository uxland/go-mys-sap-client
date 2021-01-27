package mys_sap_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
)

//Holds authentication info to send to SAP
type SAPAuth struct {
	//Authentication type in Value
	Type SAPAuthenticationType
	//A SAP Cookie or username:pwd in base64
	Value string

	Cookies map[string]string
}

const sapAuthKey = "sap-auth"

func AuthenticationMiddlewareFactory(apiSecret string) func(next http.Handler) http.Handler {
	apiSecretBytes := []byte(apiSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Header["Authorization"]
			if err != true || c == nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("error: no one token authorization exists"))
				return
			}
			var user *SAPAuth = nil
			tokenSplit := strings.Split(c[0], " ")
			switch tokenSplit[0] {
			case "Bearer":
				token, tokenError := validateToken(tokenSplit[1], apiSecretBytes)
				if tokenError != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("error: you should be authorized for this action"))
					return
				}
				cookie, _ := getCookies(token)
				user = &SAPAuth{Value: cookie, Type: Cookies}

			case "Basic":
				user = &SAPAuth{Value: tokenSplit[1], Type: Basic}
				break
			}
			bytes, _ := json.Marshal(user)
			ctx := context.WithValue(r.Context(), sapAuthKey, bytes)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func GetRequestUser(r *http.Request) *SAPAuth {
	jsonUser, _ := r.Context().Value(sapAuthKey).([]byte)
	user := &SAPAuth{}
	_ = json.Unmarshal(jsonUser, user)
	return user
}

func getCookies(cookies jwt.MapClaims) (string, error) {
	const userCtxCookieKey = "sap-usercontext"
	const ssoCookieKey = "MYSAPSSO2"
	const sapSessionCookieKey = "SAP_SESSIONID_BID_100"

	var cookie string
	for key, value := range cookies {
		if key == "accesses" {
			userContext := value.(interface{}).(map[string]interface{})[userCtxCookieKey]
			ssoCookie := value.(interface{}).(map[string]interface{})[ssoCookieKey]
			sapSession := value.(interface{}).(map[string]interface{})[sapSessionCookieKey]
			cookie = fmt.Sprintf(`%s=%s; %s=%s; %s=%s;`,
				userCtxCookieKey, userContext,
				ssoCookie, ssoCookie,
				sapSession, sapSession)
		}
	}
	return cookie, nil
}
func buildSapCookies(cookies map[string]string) string {
	var cookie string
	for key, value := range cookies {
		cookie = fmt.Sprintf("%s%s=%s; ", cookie, key, value)
	}
	return cookie
}
func validateToken(token string, apiSecret []byte) (jwt.MapClaims, error) {
	var verifiedToken, tokenError = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return apiSecret, nil
	})
	if tokenError != nil {
		return nil, errors.New("error verifying token")
	}
	claims := verifiedToken.Claims.(jwt.MapClaims)
	return claims, nil
}
