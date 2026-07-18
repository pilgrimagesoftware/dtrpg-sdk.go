// This file targets www.drivethrurpg.com — not api.drivethrurpg.com. It wraps the two
// website login endpoints that DriveThruRPG's own login page uses to turn an email/password
// pair into an application key.
//
// This is distinct from Authenticate (key_exchange.go), which exchanges an application key
// for a short-lived JWT against api.drivethrurpg.com. LoginWithCredentials produces the
// application key that Authenticate then exchanges for a session token; the two flows are
// complementary, not overlapping. Authenticate is unaffected by this file.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/pilgrimagesoftware/dtrpg-sdk.go/library"
)

// websiteBaseURL is the base URL for the DriveThruRPG website login endpoints.
const websiteBaseURL = "https://www.drivethrurpg.com"

// validateLoginResponse is the typed response from POST /validate_login_credentials.php.
//
// The endpoint returns a bare JSON array (not an object). Field order per
// dtrpg-api/LOGIN.md: [field_name, ok, message, locked].
//
// Example: ["password", true, "Locked", true]
type validateLoginResponse struct {
	FieldName string
	OK        bool
	Message   string
	Locked    bool
}

func (v *validateLoginResponse) UnmarshalJSON(data []byte) error {
	var raw [4]json.RawMessage
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 4 {
		return fmt.Errorf("expected a 4-element JSON array [field_name, ok, message, locked], got %d elements", len(arr))
	}
	copy(raw[:], arr)

	if err := json.Unmarshal(raw[0], &v.FieldName); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &v.OK); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[2], &v.Message); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[3], &v.Locked); err != nil {
		return err
	}
	return nil
}

// createAccountAppResponse is the typed response from POST /create_account_app.php.
type createAccountAppResponse struct {
	Status  string `json:"status"`
	Message struct {
		Key string `json:"key"`
	} `json:"message"`
}

// LoginWithCredentials exchanges an email/password pair for a DriveThruRPG application
// key.
//
// Calls POST /validate_login_credentials.php on www.drivethrurpg.com. If credentials are
// valid, calls POST /create_account_app.php and returns the application key from
// message.key. Both requests use multipart/form-data with email_address and password
// fields, per dtrpg-api/LOGIN.md.
//
// config is currently unused: this function always targets www.drivethrurpg.com, since the
// website login endpoints live on a separate origin from the api.drivethrurpg.com endpoint
// that library.Config describes. It is kept in the signature for symmetry with Authenticate
// and to leave room for a configurable website origin later without a breaking API change.
func LoginWithCredentials(ctx context.Context, email, password string, _ library.Config) (string, error) {
	return doLogin(ctx, email, password, websiteBaseURL)
}

func doLogin(ctx context.Context, email, password, baseURL string) (string, error) {
	// Step 1: validate credentials.
	validateURL := baseURL + "/validate_login_credentials.php"
	validateBody, err := postMultipart(ctx, validateURL, email, password)
	if err != nil {
		return "", err
	}

	var validated validateLoginResponse
	if err := json.Unmarshal(validateBody.body, &validated); err != nil {
		return "", &library.DecodeError{
			URL:     validateURL,
			Status:  validateBody.status,
			Cause:   err,
			Payload: truncatedPayload(validateBody.body),
		}
	}

	if !validated.OK {
		return "", library.ErrInvalidCredentials
	}

	// Step 2: request the application key.
	keyURL := baseURL + "/create_account_app.php"
	keyBody, err := postMultipart(ctx, keyURL, email, password)
	if err != nil {
		return "", err
	}

	var keyResult createAccountAppResponse
	if err := json.Unmarshal(keyBody.body, &keyResult); err != nil {
		return "", &library.DecodeError{
			URL:     keyURL,
			Status:  keyBody.status,
			Cause:   err,
			Payload: truncatedPayload(keyBody.body),
		}
	}

	if keyResult.Status != "success" {
		return "", &library.ApplicationKeyRequestError{Status: keyResult.Status}
	}

	return keyResult.Message.Key, nil
}

type httpResponseBody struct {
	status int
	body   []byte
}

func postMultipart(ctx context.Context, url, email, password string) (httpResponseBody, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("email_address", email); err != nil {
		return httpResponseBody{}, err
	}
	if err := writer.WriteField("password", password); err != nil {
		return httpResponseBody{}, err
	}
	if err := writer.Close(); err != nil {
		return httpResponseBody{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return httpResponseBody{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return httpResponseBody{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpResponseBody{}, err
	}
	return httpResponseBody{status: resp.StatusCode, body: body}, nil
}

func truncatedPayload(body []byte) string {
	if len(body) > logPayloadLimit {
		return string(body[:logPayloadLimit]) + "... (truncated)"
	}
	return string(body)
}
