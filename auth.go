package mys_sap_client

import (
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
}

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
			r.Header.Set("x-auth", string(bytes))
			next.ServeHTTP(w, r)
		})
	}
}

func GetRequestUser(r *http.Request) *SAPAuth {
	jsonUser := r.Header.Get("x-auth")
	user := &SAPAuth{}
	_ = json.Unmarshal([]byte(jsonUser), user)
	return user
}

func getCookies(cookies jwt.MapClaims) (string, error) {
	const UserCtxCookie = "sap-usercontext"
	const SsoCookie = "MYSAPSSO2"
	const SapSession = "SAP_SESSIONID_BID_100"
	var cookie string
	for key, value := range cookies {
		if key == "accesses" {
			userContext := value.(interface{}).(map[string]interface{})[UserCtxCookie]
			ssoCookie := value.(interface{}).(map[string]interface{})[SsoCookie]
			sapSession := value.(interface{}).(map[string]interface{})[SapSession]
			cookie = fmt.Sprintf(`%s=%s; %s=%s; %s=%s;`,
				UserCtxCookie, userContext,
				SsoCookie, ssoCookie,
				SapSession, sapSession)
		}
	}
	return cookie, nil
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
