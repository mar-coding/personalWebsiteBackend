package middlewares

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mar-coding/personalWebsiteBackend/pkg/acl"
	"github.com/mar-coding/personalWebsiteBackend/pkg/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"net/http/pprof"
	"path"
	"strings"
)

type JwtHandler struct {
	serviceCode      int32
	publicSecretKey  string
	privateSecretKey string
}

// NewJwtHandler create middleware for http handler to auth with jwt
func NewJwtHandler(serviceCode int32, publicSecretKey, privateSecretKey string) *JwtHandler {
	return &JwtHandler{
		serviceCode:      serviceCode,
		publicSecretKey:  publicSecretKey,
		privateSecretKey: privateSecretKey,
	}
}

func (j *JwtHandler) SetAuthToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		r = r.WithContext(context.WithValue(r.Context(), "Authorization", authorization))
		next.ServeHTTP(w, r)
	})
}

func (j *JwtHandler) JwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		token := strings.Split(authorization, " ")
		if len(token) < 2 {
			http.Error(w, "jwt token is invalid", 401)
			return
		}

		if token[1] == "" {
			http.Error(w, "jwt token not found in header", 401)
			return
		}

		if token[0] != "Bearer" {
			http.Error(w, "missing prefix Bearer in your token", 401)
			return
		}

		aclController, err := acl.NewWithJwt(j.serviceCode, token[1],
			acl.WithPublicSecretKey(j.publicSecretKey),
			acl.WithPrivateSecretKey(j.privateSecretKey),
		)
		if err != nil {
			if errors.Is(err, acl.ErrSecretKeyIsEmpty) {
				http.Error(w, "public secret key is empty", 500)
				return
			}
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "jwt token expired", 401)
				return
			}
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				http.Error(w, "jwt token signature is invalid", 401)
				return
			}
			http.Error(w, "jwt token is invalid", 401)
			return
		}

		payload, err := jwt.ParseUnverifiedJwtToken(token[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		if !aclController.HasAnyPermissionsAccess(payload.Permissions[j.serviceCode]...) {
			http.Error(w, "you don't have permission to access method", 403)
			return
		}

		r = r.WithContext(aclController.SetAclToContext(r.Context()))
		next.ServeHTTP(w, r)
	})
}

// ErrorHandler convert grpc status error to http error
func ErrorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	runtime.DefaultHTTPErrorHandler(ctx, mux, m, w, r, err)
}

// AllowCORS add cors to http handler
func AllowCORS(h http.Handler, origins []string, customHeaders ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if len(origins) != 0 {
			if origin == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if !checkOrigin(origin, origins) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		} else {
			origin = "*"
		}

		headers := []string{
			"Accept",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"ResponseType",
		}

		headers = append(headers, customHeaders...)

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// SetRuntimeAsRootHandler set runtime mux as root handler http server mux
func SetRuntimeAsRootHandler(mux *http.ServeMux, rMux *runtime.ServeMux) *http.ServeMux {
	mux.Handle("/", rMux)
	return mux
}

// SwaggerHandler add swagger file embedded to http handler path, swaggerFileName (swagger.json or swagger.yaml and etc)
func SwaggerHandler(mux *http.ServeMux, swaggerFileName string, swagger []byte) *http.ServeMux {
	mux.HandleFunc("/"+swaggerFileName, func(w http.ResponseWriter, _ *http.Request) {
		w.Write(swagger)
	})
	return mux
}

// DebuggerHandler add pprof handlers handlers to http server mux
func DebuggerHandler(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())
	return mux
}

// GrpcGatewayUploadFileHandler help for upload file by grpc gateway
func GrpcGatewayUploadFileHandler(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
				newR, err := createRequestFromMultiPart(r)
				if err != nil {
					w.WriteHeader(400)
					return
				}
				otherHandler.ServeHTTP(w, newR)
			} else {
				otherHandler.ServeHTTP(w, r)
			}
		}
	})
}

// SessionHandler add sessionToken to GRPC metadata
func SessionHandler(ctx context.Context, r *http.Request) metadata.MD {
	sessionToken := ""
	session, err := r.Cookie(SESSION_COOKIE_KEY)
	if err == nil {
		sessionToken = session.Value
	}
	return metadata.Pairs(SESSION_COOKIE_KEY, sessionToken)
}

func getFormFile(r *http.Request, name string) ([]byte, error) {
	file, _, err := r.FormFile(name)
	if err != nil {
		return nil, fmt.Errorf("not found")
	}
	defer file.Close()
	buf := bytes.Buffer{}
	io.Copy(&buf, file)
	if err != nil {
		return nil, fmt.Errorf("error while reading form file")
	}
	return buf.Bytes(), nil
}

func expandMacros(jsonData map[string]interface{}, r *http.Request, res *map[string]interface{}) error {
	for name, val := range jsonData {
		if inner, isObject := val.(map[string]interface{}); isObject {
			(*res)[name] = map[string]interface{}{}
			resInner, _ := (*res)[name].(map[string]interface{})
			err := expandMacros(inner, r, &resInner)
			if err != nil {
				return err
			}
		} else if valStr, ok := val.(string); ok && strings.HasPrefix(valStr, "$") {
			data, err := getFormFile(r, valStr)
			if err != nil {
				return fmt.Errorf("can not get file %s", val.(string))
			}
			(*res)[name] = base64.StdEncoding.EncodeToString(data)
		} else {
			(*res)[name] = val
		}
	}
	return nil
}

func createRequestFromMultiPart(r *http.Request) (*http.Request, error) {
	json_data, err := getFormFile(r, "data")
	if err != nil {
		return nil, err
	}
	rawJson := json.RawMessage(json_data)
	var decoded map[string]interface{}
	expanded := map[string]interface{}{}
	json.Unmarshal(rawJson, &decoded)
	err = expandMacros(decoded, r, &expanded)
	str, _ := json.Marshal(expanded)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(str)
	newR, err := http.NewRequest(http.MethodPost, r.URL.String(), reader)
	if err != nil {
		return nil, err
	}
	return newR, nil
}

func checkOrigin(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || originMatchesPattern(origin, allowedOrigin) {
			return true
		}
	}
	return false
}

func originMatchesPattern(origin, pattern string) bool {
	matched, err := path.Match(pattern, origin)
	return err == nil && matched
}
