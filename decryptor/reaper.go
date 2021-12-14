package decryptor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// ReaperDecryptor will call out to the reaper service to decrypt arguments
// passed into a BizApp.  It will use a one-off signing key to handle authorizing
// a request to decrypt arguments associated with a particular CommandExecutor.  The
// reaper will return a list of decrypted arguments, given a list of arguments in the
// body of the request
type ReaperDecryptor struct {
	BaseURL           string
	SigningKey        []byte
	CommandExecutorID string
}

// Authorization uses a signed nonce, so we really just
// want to sign a string as the token we use for authorization.  In order
// for it to be signed correctly, go-jwt requires us to implement a Valid()
// method.
type tokenClaims string

func (t tokenClaims) Valid() error {
	return nil
}

//
// {"data": {"arguments": ["--arg1=zzz", "--arg2=bbb"]}}
//
type reaperResponse struct {
	Data struct {
		Arguments []string `json:"arguments"`
	} `json:"data"`
}

const vaultEncryptStart = "OC_ENCRYPTED"
const vaultEncryptEnd = "DETPYRCNE_CO"
const decryptPath = "/runner/decrypt_arguments"

var vaultRegex = regexp.MustCompile(vaultEncryptStart + "(.*)" + vaultEncryptEnd)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// NewReaperDecryptor returns a Decryptor implementation that will call out to
// the reaper service to decrypt any encrypted arguments.
func NewReaperDecryptor(url, signingKey, commandExecutorID string) Decryptor {
	return &ReaperDecryptor{
		BaseURL:           url,
		SigningKey:        []byte(signingKey),
		CommandExecutorID: commandExecutorID,
	}
}

// DecryptArguments replaces any encrypted values with their decrypted values
// using the reaper service.
func (r *ReaperDecryptor) DecryptArguments(args []string) ([]string, error) {
	// If no encrypted arguments are found, do not attempt to decrypt
	shouldDecrypt := false
	for _, arg := range args {
		if vaultRegex.MatchString(arg) {
			shouldDecrypt = true
			break
		}
	}

	if !shouldDecrypt {
		return args, nil
	}

	payload := make(map[string]interface{})
	payload["commandExecutorId"] = r.CommandExecutorID
	payload["arguments"] = args
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return args, fmt.Errorf("error creating payload to send for decryption: %s", err)
	}
	payloadBody := bytes.NewBuffer(payloadBytes)

	reqURL := fmt.Sprintf("%s%s", r.BaseURL, decryptPath)
	retryReq, err := retryablehttp.NewRequest("POST", reqURL, payloadBody)
	if err != nil {
		return args, fmt.Errorf("error building request to send for decryption: %s", err)
	}

	cl := defaultClientWithRetries()
	addJWTHeader(retryReq, r.SigningKey)

	resp, err := cl.Do(retryReq)
	if err != nil {
		return args, fmt.Errorf("error response from decryption service: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		io.Copy(ioutil.Discard, resp.Body)
		return args, fmt.Errorf("decryption service responsed with status code %d", resp.StatusCode)
	}

	var respObj reaperResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&respObj)
	if err != nil {
		return args, fmt.Errorf("error parsing response from decryption service: %s", err)
	}

	return respObj.Data.Arguments, err
}

func addJWTHeader(req *retryablehttp.Request, signingKey []byte) error {
	nonce := generateNonce()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, nonce)

	signed, err := token.SignedString(signingKey)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signed))

	return err
}

// random string of 32 bytes to be signed into a JWT - value is not used
func generateNonce() tokenClaims {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return tokenClaims(b)
}

// Using a HTTP client that will automatically retry 5xx errors to ensure that our connection
// is resilient
func defaultClientWithRetries() *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryWaitMin = 3 * time.Second
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil {
			return true, err
		}

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			return true, err
		}

		if err != nil || ctx.Err() != nil {
			return true, err
		}
		return false, err
	}

	retryClient.RequestLogHook = func(l retryablehttp.Logger, req *http.Request, retryCount int) {
		*req = *req.Clone(context.TODO())
	}

	return retryClient
}
