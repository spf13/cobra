package decryptor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecryptArguments_NoneEncrypted(t *testing.T) {
	dec := NewReaperDecryptor("http://url.com/not/used", string(generateNonce()), "1234")
	firstArg := "--arg1=aaa"
	secondArg := "--zzz=bbb"
	args := []string{firstArg, secondArg}

	decryptedArgs, err := dec.DecryptArguments(args)
	if err != nil {
		t.Error("NoneEncrypted should not return an error")
	}

	if decryptedArgs[0] != firstArg {
		t.Errorf("NoneEncrypted - first arg should be %s, got %s", firstArg, decryptedArgs[0])
	}

	if decryptedArgs[1] != secondArg {
		t.Errorf("NoneEncrypted - second arg should be %s, got %s", firstArg, decryptedArgs[0])
	}
}

func TestDecryptArguments_EncryptedArgsDecoded(t *testing.T) {
	encryptedKey := "zzzzzzz"
	decryptedKey := "aaaaaaa"
	firstArg := fmt.Sprintf("--arg1=OC_ENCRYPTED%sDETPYRCNE_CO", encryptedKey)
	decryptedArg := fmt.Sprintf("--arg1=%s", decryptedKey)
	secondArg := "--arg2=iamnotencrypted"
	args := []string{firstArg, secondArg}

	reaperTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		payloadObj := make(map[string]interface{})
		dataObj := make(map[string]interface{})
		dataObj["arguments"] = []string{decryptedArg, secondArg}
		payloadObj["data"] = dataObj

		payloadBytes, err := json.Marshal(payloadObj)
		if err != nil {
			t.Errorf("error returning test data JSON: %s", err)
			t.FailNow()
		}

		b := bytes.NewBuffer(payloadBytes)
		b.WriteTo(w)
	}))
	defer reaperTestServer.Close()

	dec := NewReaperDecryptor(reaperTestServer.URL, string(generateNonce()), "1234")
	decryptedArgs, err := dec.DecryptArguments(args)
	if err != nil {
		t.Error("EncryptedArgsDecoded should not return an error")
	}

	if decryptedArgs[0] != decryptedArg {
		t.Errorf("EncryptedArgsDecoded - first arg should be %s, got %s", decryptedArg, decryptedArgs[0])
	}

	if decryptedArgs[1] != secondArg {
		t.Errorf("EncryptedArgsDecoded - second arg should be %s, got %s", firstArg, decryptedArgs[0])
	}
}
