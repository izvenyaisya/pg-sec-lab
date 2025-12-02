package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validate(&p); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &p, nil
}

func validate(p *Policy) error {
	for tableName, tp := range p.Tables {
		if tp.RLS.Enabled && tp.RLS.SelectPolicy == "" {
			return fmt.Errorf("table %s: RLS enabled but select_policy is empty", tableName)
		}

		for _, mask := range tp.Masks {
			if mask.Column == "" {
				return fmt.Errorf("table %s: mask has empty column", tableName)
			}
			if mask.Expression == "" {
				return fmt.Errorf("table %s: mask for column %s has empty expression", tableName, mask.Column)
			}
			if mask.ExposedAs == "" {
				return fmt.Errorf("table %s: mask for column %s has empty exposed_as", tableName, mask.Column)
			}
		}
	}

	return nil
}
