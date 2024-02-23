package geomys

import (
	"slices"
	"strings"
)

func CanonicalizeModuleName(moduleName string) string {
	var canonNameBuilder strings.Builder
	splitString := strings.Split(moduleName, "/")

	moduleName = splitString[0]

	splitModuleName := strings.Split(moduleName, ".")
	slices.Reverse(splitModuleName)
	canonNameBuilder.WriteString(strings.Join(splitModuleName, "_"))
	canonNameBuilder.WriteString("_")
	canonNameBuilder.WriteString(strings.Join(splitString[1:], "_"))

	canonName := canonNameBuilder.String()
	canonName = strings.Replace(canonName, "-", "_", -1)
	return canonName
}
