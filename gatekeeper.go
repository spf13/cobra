package cobra

import (
	"regexp"
)

const vaultEncryptStart = "OC_ENCRYPTED"
const vaultEncryptEnd = "DETPYRCNE_CO"

var vaultRegex = regexp.MustCompile(vaultEncryptStart + "(.*)" + vaultEncryptEnd)

type Gatekeeper struct{}

func (g *Gatekeeper) DecryptFlags(flags []string) ([]string, error) {
	var decryptedFlags []string

	for _, fl := range flags {
		keyToDecrypt := extractSecretKey(fl)
		if keyToDecrypt == "" {
			continue
		} else {
			// TODO: Gatekeeper client should perform decryption here
			// TODO: Remove recursive decryption to the Gatekeeper client
			// TODO: Move macro replacement to Gatekeeper package
			decryptedFlags = append(decryptedFlags, keyToDecrypt)
		}
	}

	return decryptedFlags, nil
}

func extractSecretKey(macro string) string {
	matches := vaultRegex.FindStringSubmatch(macro)
	if len(matches) == 2 && matches[1] != "" {
		secretKey := matches[1]
		return secretKey
	}
	return ""
}
