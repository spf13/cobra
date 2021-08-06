package cobra

import (
	"fmt"
	"os"
	"strings"
)

const (
	activeHelpMarker = "_activeHelp_ "
	// The below values should not be changed: programs will be using them explicitly
	// in their user documentation, and users will be using them explicitly.
	activeHelpEnvVarSuffix  = "_ACTIVE_HELP"
	activeHelpGlobalEnvVar  = "COBRA_ACTIVE_HELP"
	activeHelpGlobalDisable = "0"
)

// AppendActiveHelp adds the specified string to the specified array to be used as ActiveHelp.
// Such strings will be processed by the completion script and will be shown as ActiveHelp
// to the user.
// The array parameter should be the array that will contain the completions.
// This function can be called multiple times before and/or after completions are added to
// the array.  Each time this function is called with the same array, the new
// ActiveHelp line will be shown below the previous ones when completion is triggered.
func AppendActiveHelp(compArray []string, activeHelpStr string) []string {
	return append(compArray, fmt.Sprintf("%s%s", activeHelpMarker, activeHelpStr))
}

// activeHelpEnvVar returns the name of the program-specific ActiveHelp environment
// variable.  It has the format <PROGRAM>_ACTIVE_HELP where <PROGRAM> is the name of the
// root command in upper case, with all - replaced by _.
// This format should not be changed: users will be using it explicitly.
func activeHelpEnvVar(name string) string {
	activeHelpEnvVar := strings.ToUpper(fmt.Sprintf("%s%s", name, activeHelpEnvVarSuffix))
	return strings.ReplaceAll(activeHelpEnvVar, "-", "_")
}

// setActiveHelpConfig first checks the global environment variable
// of ActiveHelp to see if it is disabling active help, and if it is not,
// it then looks to the program-specific variable.
// It then sets the ActiveHelpConfig value to make it available when
// calling the completion function.  We also set it on the root,
// just in case users try to access it from there.
func setActiveHelpConfig(cmd *Command) {
	activeHelpCfg := os.Getenv(activeHelpGlobalEnvVar)
	if activeHelpCfg != activeHelpGlobalDisable {
		activeHelpCfg = os.Getenv(activeHelpEnvVar(cmd.Root().Name()))
	}

	cmd.ActiveHelpConfig = activeHelpCfg
	cmd.Root().ActiveHelpConfig = cmd.ActiveHelpConfig
}
