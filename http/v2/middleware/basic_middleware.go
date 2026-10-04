package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"
)

func BasicAuth(realm string, creds map[string]string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqUser, reqPwd, ok := r.BasicAuth()
			if !ok {
				basicAuthFailed(w, realm)
				return
			}

			pswd, found := creds[reqUser]
			if !found || subtle.ConstantTimeCompare([]byte(reqPwd), []byte(pswd)) != 1 {
				basicAuthFailed(w, realm)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func basicAuthFailed(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
