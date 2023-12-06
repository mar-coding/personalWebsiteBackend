package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mar-coding/personalWebsiteBackend/pkg/errorHandler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"net/url"
)

const (
	_defaultHeaderKey              = "x-captcha-key"
	_defaultGoogleCaptchaVerifyAPI = "https://www.google.com/recaptcha/api/siteverify"
)

type GoogleCaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs any      `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func googleCaptchaValidation(ctx context.Context, secretKey string, handler errorHandler.Handler) error {
	challengeKey, err := extractHeaderValueFromContext(ctx, _defaultHeaderKey)
	if err != nil {
		return handler.New(codes.InvalidArgument, nil, err.Error())
	}

	client := http.DefaultClient

	req, err := http.NewRequest("POST", _defaultGoogleCaptchaVerifyAPI, nil)
	if err != nil {
		return handler.New(codes.Internal, nil, "failed to create request captcha validation, got error %s", err.Error())
	}

	query := url.Values{}
	query["secret"] = []string{secretKey}
	query["response"] = []string{challengeKey}

	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return handler.New(codes.Internal, nil, "failed to do request captcha validation, got error %s", err.Error())
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return handler.New(codes.Internal, nil, "failed to read response body captcha validation, got error %s", err.Error())
	}

	response := new(GoogleCaptchaResponse)
	if err := json.Unmarshal(b, response); err != nil {
		return handler.New(codes.Internal, nil, "failed to unmarshal captcha validation, got error %s", err.Error())
	}

	if !response.Success {
		return handler.New(codes.PermissionDenied, nil, "your challenge is unsuccessful, try again to complete captcha challenge")
	}

	return nil
}

func extractHeaderValueFromContext(ctx context.Context, header string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("failed to get metadata from context")
	}

	foundedHeaders, ok := md[header]
	if !ok {
		return "", errors.New("failed to find header key X-Captcha-Key in metadata")
	}

	return foundedHeaders[0], nil
}
