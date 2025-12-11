package env

import (
	"os"
	"strings"
)

// GetShellVariables returns environment variables from the current shell
func GetShellVariables() []Variable {
	var vars []Variable

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		vars = append(vars, Variable{
			Key:      key,
			Value:    value,
			Source:   SourceShell,
			IsSecret: IsSecret(key),
			Line:     -1,
		})
	}

	return vars
}
