package utils

import (
	"fmt"
	"strings"

	"github.com/bazelbuild/buildtools/warn"
)

func ValidateModes(inputType, mode, lint *string, dflag *bool) error {
	// Check input type.
	switch *inputType {
	case "build", "bzl", "workspace", "default", "auto":
		// ok

	default:
		return fmt.Errorf("buildifier: unrecognized input type %s; valid types are build, bzl, workspace, default, auto\n", *inputType)
	}

	if *dflag {
		if *mode != "" {
			return fmt.Errorf("buildifier: cannot specify both -d and -mode flags\n")
		}
		*mode = "diff"
	}

	// Check mode.
	switch *mode {
	case "":
		*mode = "fix"

	case "check", "diff", "fix", "print_if_changed":
		// ok

	default:
		return fmt.Errorf("buildifier: unrecognized mode %s; valid modes are check, diff, fix\n", *mode)
	}

	// Check lint mode.
	switch *lint {
	case "":
		*lint = "off"

	case "off", "warn":
		// ok

	case "fix":
		if *mode != "fix" {
			return fmt.Errorf("buildifier: --lint=fix is only compatible with --mode=fix\n")
		}

	default:
		return fmt.Errorf("buildifier: unrecognized lint mode %s; valid modes are warn and fix\n", *lint)
	}

	return nil
}

func ValidateWarnings(warnings *string) ([]string, error) {

	// Check lint warnings
	var warningsList []string
	switch *warnings {
	case "", "default":
		warningsList = warn.DefaultWarnings
	case "all":
		warningsList = warn.AllWarnings
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
			return []string{}, fmt.Errorf("buildifier: warning categories with modifiers (\"+\" or \"-\") can't me mixed with raw warning categories\n")
		}
		if len(warningsList) == 0 {
			for _, warning := range warn.DefaultWarnings {
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
