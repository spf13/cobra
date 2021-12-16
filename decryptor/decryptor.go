package decryptor

import "os"

type Decryptor interface {
	DecryptArguments([]string) ([]string, error)
}

func NewDecryptor() Decryptor {
	if IsCloudRunner() && ReaperURL() != "" && BizAppAuthToken() != "" && CommandExecutorID() != "" {
		return NewReaperDecryptor(ReaperURL(), BizAppAuthToken(), CommandExecutorID())
	}

	return NewNoopDecryptor()
}

func IsCloudRunner() bool {
	return os.Getenv("OC_CLOUDRUNNER_CONFIG") != ""
}

func ReaperURL() string {
	return os.Getenv("REAPER_URL")
}

func BizAppAuthToken() string {
	return os.Getenv("BIZ_APP_AUTH_TOKEN")
}

func CommandExecutorID() string {
	return os.Getenv("OC_COMMAND_EXECUTOR_ID")
}
