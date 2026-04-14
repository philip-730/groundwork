package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	tmpl "github.com/philip-730/groundwork/internal/template"
)

// ParseVars converts a slice of "key=value" strings (from --var flags) into a
// map. The value is everything after the first '=', so values may contain '='.
func ParseVars(vars []string) (map[string]string, error) {
	out := make(map[string]string, len(vars))
	for _, v := range vars {
		idx := strings.IndexByte(v, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid --var %q: expected key=value", v)
		}
		key := v[:idx]
		val := v[idx+1:]
		if strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("invalid --var %q: key must not be empty", v)
		}
		out[key] = val
	}
	return out, nil
}

// CollectInputs returns a complete map of input values for the template.
// Values already present in provided are used as-is. Missing required inputs
// are read interactively from r (one line per input). Missing optional inputs
// fall back to their declared default. Returns an error if a required input
// cannot be satisfied.
func CollectInputs(t *tmpl.Template, provided map[string]string, r io.Reader, w io.Writer) (map[string]string, error) {
	result := make(map[string]string, len(t.Inputs))

	// Copy already-provided values first.
	for k, v := range provided {
		result[k] = v
	}

	scanner := bufio.NewScanner(r)

	for name, spec := range t.Inputs {
		if _, ok := result[name]; ok {
			continue // already set via --var
		}

		if spec.Required {
			val, err := prompt(scanner, w, name, spec, true)
			if err != nil {
				return nil, err
			}
			result[name] = val
			continue
		}

		if spec.Default != "" {
			// Optional with a default — prompt but show the default.
			val, err := prompt(scanner, w, name, spec, false)
			if err != nil {
				return nil, err
			}
			if val == "" {
				val = spec.Default
			}
			result[name] = val
		}
		// Optional with no default: omit from result (template must handle absence).
	}

	return result, nil
}

func prompt(scanner *bufio.Scanner, w io.Writer, name string, spec tmpl.Input, required bool) (string, error) {
	for {
		if spec.Default != "" {
			fmt.Fprintf(w, "  %s (%s, default: %s): ", name, spec.Type, spec.Default)
		} else if required {
			fmt.Fprintf(w, "  %s (%s, required): ", name, spec.Type)
		} else {
			fmt.Fprintf(w, "  %s (%s): ", name, spec.Type)
		}

		if !scanner.Scan() {
			if required {
				return "", fmt.Errorf("input %q is required but stdin closed before a value was given", name)
			}
			return "", nil
		}

		val := strings.TrimSpace(scanner.Text())
		if val != "" || !required {
			return val, nil
		}
		fmt.Fprintln(w, "  (value required, please try again)")
	}
}
