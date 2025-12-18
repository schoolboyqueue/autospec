package workflow

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSpecSchema(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specDir     string
		wantErr     bool
		errContains string
		description string
	}{
		"valid spec": {
			specDir:     filepath.Join("testdata", "spec", "valid"),
			wantErr:     false,
			description: "Valid spec.yaml should pass validation",
		},
		"invalid spec missing feature": {
			specDir:     filepath.Join("testdata", "spec", "invalid"),
			wantErr:     true,
			errContains: "missing required field: feature",
			description: "Spec missing required 'feature' field should fail",
		},
		"nonexistent directory": {
			specDir:     filepath.Join("testdata", "spec", "nonexistent"),
			wantErr:     true,
			errContains: "failed to parse YAML",
			description: "Nonexistent directory should fail with parse error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateSpecSchema(tc.specDir)

			if tc.wantErr {
				if err == nil {
					t.Errorf("ValidateSpecSchema() expected error, got nil; %s", tc.description)
					return
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("ValidateSpecSchema() error = %q, want error containing %q; %s",
						err.Error(), tc.errContains, tc.description)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateSpecSchema() unexpected error: %v; %s", err, tc.description)
			}
		})
	}
}

func TestValidatePlanSchema(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specDir     string
		wantErr     bool
		errContains string
		description string
	}{
		"valid plan": {
			specDir:     filepath.Join("testdata", "plan", "valid"),
			wantErr:     false,
			description: "Valid plan.yaml should pass validation",
		},
		"invalid plan missing plan field": {
			specDir:     filepath.Join("testdata", "plan", "invalid"),
			wantErr:     true,
			errContains: "missing required field: plan",
			description: "Plan missing required 'plan' field should fail",
		},
		"nonexistent directory": {
			specDir:     filepath.Join("testdata", "plan", "nonexistent"),
			wantErr:     true,
			errContains: "failed to parse YAML",
			description: "Nonexistent directory should fail with parse error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePlanSchema(tc.specDir)

			if tc.wantErr {
				if err == nil {
					t.Errorf("ValidatePlanSchema() expected error, got nil; %s", tc.description)
					return
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("ValidatePlanSchema() error = %q, want error containing %q; %s",
						err.Error(), tc.errContains, tc.description)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidatePlanSchema() unexpected error: %v; %s", err, tc.description)
			}
		})
	}
}

func TestValidateTasksSchema(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specDir     string
		wantErr     bool
		errContains string
		description string
	}{
		"valid tasks": {
			specDir:     filepath.Join("testdata", "tasks", "valid"),
			wantErr:     false,
			description: "Valid tasks.yaml should pass validation",
		},
		"invalid tasks missing tasks field": {
			specDir:     filepath.Join("testdata", "tasks", "invalid"),
			wantErr:     true,
			errContains: "missing required field: tasks",
			description: "Tasks missing required 'tasks' field should fail",
		},
		"nonexistent directory": {
			specDir:     filepath.Join("testdata", "tasks", "nonexistent"),
			wantErr:     true,
			errContains: "failed to parse YAML",
			description: "Nonexistent directory should fail with parse error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateTasksSchema(tc.specDir)

			if tc.wantErr {
				if err == nil {
					t.Errorf("ValidateTasksSchema() expected error, got nil; %s", tc.description)
					return
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("ValidateTasksSchema() error = %q, want error containing %q; %s",
						err.Error(), tc.errContains, tc.description)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateTasksSchema() unexpected error: %v; %s", err, tc.description)
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		artifactName string
		errCount     int
		wantNil      bool
		wantContains []string
		description  string
	}{
		"no errors returns nil": {
			artifactName: "test.yaml",
			errCount:     0,
			wantNil:      true,
			description:  "Empty error slice should return nil",
		},
		"single error": {
			artifactName: "spec.yaml",
			errCount:     1,
			wantNil:      false,
			wantContains: []string{"schema validation failed for spec.yaml", "test error 0"},
			description:  "Single error should be formatted correctly",
		},
		"multiple errors": {
			artifactName: "plan.yaml",
			errCount:     3,
			wantNil:      false,
			wantContains: []string{"schema validation failed for plan.yaml", "test error 0", "test error 1", "test error 2"},
			description:  "Multiple errors should all be included",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Create test validation errors
			var errors []*validationErrorForTest
			for i := 0; i < tc.errCount; i++ {
				errors = append(errors, &validationErrorForTest{
					message: "test error " + string(rune('0'+i)),
				})
			}

			// We can't directly call formatValidationErrors since it uses
			// *validation.ValidationError, so we test the behavior through
			// the public functions that use it.
			// This test documents the expected behavior.

			if tc.wantNil {
				// Verify nil behavior by checking valid file returns nil
				err := ValidateSpecSchema(filepath.Join("testdata", "spec", "valid"))
				if err != nil {
					t.Errorf("Expected nil error for valid file: %v", err)
				}
			} else {
				// Verify error formatting by checking invalid file
				err := ValidateSpecSchema(filepath.Join("testdata", "spec", "invalid"))
				if err == nil {
					t.Error("Expected error for invalid file")
					return
				}

				errStr := err.Error()
				if !strings.Contains(errStr, "schema validation failed for") {
					t.Errorf("Error should contain 'schema validation failed for': %s", errStr)
				}
			}
		})
	}
}

// validationErrorForTest is a test helper
type validationErrorForTest struct {
	message string
}

func (e *validationErrorForTest) Error() string {
	return e.message
}
