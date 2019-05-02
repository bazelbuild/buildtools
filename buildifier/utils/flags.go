package utils

import (
	"fmt"
	"strings"
)

// ValidateInputType validates the value of --type
func ValidateInputType(inputType *string) error {
	switch *inputType {
	case "build", "bzl", "workspace", "default", "auto":
		return nil

	default:
		return fmt.Errorf("unrecognized input type %s; valid types are build, bzl, workspace, default, auto", *inputType)
	}
}

// ValidateFormat validates the value of --format
func ValidateFormat(format, mode *string) error {
	switch *format {
	case "":
		return nil

	case "text", "json":
		if *mode != "check" {
			return fmt.Errorf("cannot specify --format without --type=check")
		}

	default:
		return fmt.Errorf("unrecognized format %s; valid types are text, json", *format)
	}
	return nil
}

// isRecognizedMode checks whether the given mode is one of the valid modes.
func isRecognizedMode(validModes []string, mode string) bool {
	for _, m := range validModes {
		if mode == m {
			return true
		}
	}
	return false
}

// ValidateModes validates flags --mode, --lint, and -d
func ValidateModes(mode, lint *string, dflag *bool, additionalModes ...string) error {
	if *dflag {
		if *mode != "" {
			return fmt.Errorf("cannot specify both -d and -mode flags")
		}
		*mode = "diff"
	}

	// Check mode.
	validModes := []string{"check", "diff", "fix", "print_if_changed"}
	validModes = append(validModes, additionalModes...)

	if *mode == "" {
		*mode = "fix"
	} else if !isRecognizedMode(validModes, *mode) {
		return fmt.Errorf("unrecognized mode %s; valid modes are %s", *mode, strings.Join(validModes, ", "))
	}

	// Check lint mode.
	switch *lint {
	case "":
		*lint = "off"

	case "off", "warn":
		// ok

	case "fix":
		if *mode != "fix" {
			return fmt.Errorf("--lint=fix is only compatible with --mode=fix")
		}

	default:
		return fmt.Errorf("unrecognized lint mode %s; valid modes are warn and fix", *lint)
	}

	return nil
}

// ValidateWarnings validates the value of the --warnings flag
func ValidateWarnings(warnings *string, allWarnings, defaultWarnings *[]string) ([]string, error) {

	// Check lint warnings
	var warningsList []string
	switch *warnings {
	case "", "default":
		warningsList = *defaultWarnings
	case "all":
		warningsList = *allWarnings
	default:
		// Either all or no warning categories should start with "+" or "-".
		// If all of them start with "+" or "-", the semantics is
		// "default set of warnings + something - something".
		plus := map[string]bool{}
		minus := map[string]bool{}
		for _, warning := range strings.Split(*warnings, ",") {
			if strings.HasPrefix(warning, "+") {
				plus[warning[1:]] = true
			} else if strings.HasPrefix(warning, "-") {
				minus[warning[1:]] = true
			} else {
				warningsList = append(warningsList, warning)
			}
		}
		if len(warningsList) > 0 && (len(plus) > 0 || len(minus) > 0) {
			return []string{}, fmt.Errorf("warning categories with modifiers (\"+\" or \"-\") can't me mixed with raw warning categories")
		}
		if len(warningsList) == 0 {
			for _, warning := range *defaultWarnings {
				if !minus[warning] {
					warningsList = append(warningsList, warning)
				}
			}
			for warning := range plus {
				warningsList = append(warningsList, warning)
			}
		}
	}
	return warningsList, nil
}
