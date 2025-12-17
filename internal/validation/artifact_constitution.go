package validation

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ConstitutionValidator validates constitution.yaml artifacts.
type ConstitutionValidator struct {
	baseValidator
}

// Type returns the artifact type.
func (v *ConstitutionValidator) Type() ArtifactType {
	return ArtifactTypeConstitution
}

// Validate validates a constitution.yaml file at the given path.
func (v *ConstitutionValidator) Validate(path string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Parse the YAML file
	root, err := parseYAMLFile(path)
	if err != nil {
		result.AddError(&ValidationError{
			Path:    path,
			Message: fmt.Sprintf("failed to parse YAML: %v", err),
			Hint:    "Check the YAML syntax for errors",
		})
		return result
	}

	rootMapping := getRootMapping(root)
	if rootMapping == nil {
		result.AddError(&ValidationError{
			Path:    path,
			Message: "expected a YAML mapping at document root",
			Hint:    "The constitution.yaml file should start with key-value pairs, not a list or scalar",
		})
		return result
	}

	// Validate required fields
	constitutionNode := validateRequiredField(rootMapping, "constitution", result)
	principlesNode := validateRequiredField(rootMapping, "principles", result)

	// Validate constitution section
	if constitutionNode != nil {
		v.validateConstitutionSection(constitutionNode, result)
	}

	// Validate principles section
	if principlesNode != nil {
		v.validatePrinciples(principlesNode, result)
	}

	// Validate optional sections
	sectionsNode := findNode(rootMapping, "sections")
	if sectionsNode != nil {
		v.validateSections(sectionsNode, result)
	}

	// Build summary if valid
	if result.Valid {
		result.Summary = v.buildSummary(rootMapping)
	}

	return result
}

// validateConstitutionSection validates the constitution section.
func (v *ConstitutionValidator) validateConstitutionSection(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "constitution", yaml.MappingNode, "object", result) {
		return
	}

	// Required fields in constitution
	validateRequiredField(node, "project_name", result)
	validateRequiredField(node, "version", result)
}

// validatePrinciples validates the principles section.
func (v *ConstitutionValidator) validatePrinciples(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "principles", yaml.SequenceNode, "array", result) {
		return
	}

	for i, principleNode := range node.Content {
		path := fmt.Sprintf("principles[%d]", i)
		v.validatePrinciple(principleNode, path, result)
	}
}

// validatePrinciple validates a single principle.
func (v *ConstitutionValidator) validatePrinciple(node *yaml.Node, path string, result *ValidationResult) {
	if node.Kind != yaml.MappingNode {
		result.AddError(&ValidationError{
			Path:     path,
			Line:     getNodeLine(node),
			Message:  fmt.Sprintf("wrong type for '%s'", path),
			Expected: "object",
			Actual:   nodeKindToString(node.Kind),
		})
		return
	}

	// Required fields
	requiredFields := []string{"name", "id", "priority", "description"}
	for _, field := range requiredFields {
		fieldNode := findNode(node, field)
		if fieldNode == nil {
			result.AddError(&ValidationError{
				Path:    fmt.Sprintf("%s.%s", path, field),
				Line:    getNodeLine(node),
				Message: fmt.Sprintf("missing required field: %s", field),
				Hint:    fmt.Sprintf("Add the '%s' field to this principle", field),
			})
		}
	}

	// Validate priority enum
	priorityNode := findNode(node, "priority")
	if priorityNode != nil {
		validateEnumValue(priorityNode, path+".priority",
			[]string{"NON-NEGOTIABLE", "MUST", "SHOULD", "MAY"}, result)
	}

	// Validate category enum if present
	categoryNode := findNode(node, "category")
	if categoryNode != nil {
		validateEnumValue(categoryNode, path+".category",
			[]string{"quality", "architecture", "process", "security"}, result)
	}
}

// validateSections validates the sections array.
func (v *ConstitutionValidator) validateSections(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "sections", yaml.SequenceNode, "array", result) {
		return
	}

	for i, sectionNode := range node.Content {
		path := fmt.Sprintf("sections[%d]", i)
		v.validateSection(sectionNode, path, result)
	}
}

// validateSection validates a single section.
func (v *ConstitutionValidator) validateSection(node *yaml.Node, path string, result *ValidationResult) {
	if node.Kind != yaml.MappingNode {
		result.AddError(&ValidationError{
			Path:     path,
			Line:     getNodeLine(node),
			Message:  fmt.Sprintf("wrong type for '%s'", path),
			Expected: "object",
			Actual:   nodeKindToString(node.Kind),
		})
		return
	}

	// Required fields
	requiredFields := []string{"name", "content"}
	for _, field := range requiredFields {
		fieldNode := findNode(node, field)
		if fieldNode == nil {
			result.AddError(&ValidationError{
				Path:    fmt.Sprintf("%s.%s", path, field),
				Line:    getNodeLine(node),
				Message: fmt.Sprintf("missing required field: %s", field),
				Hint:    fmt.Sprintf("Add the '%s' field to this section", field),
			})
		}
	}
}

// buildSummary builds the summary for a valid constitution artifact.
func (v *ConstitutionValidator) buildSummary(root *yaml.Node) *ArtifactSummary {
	summary := &ArtifactSummary{
		Type:   ArtifactTypeConstitution,
		Counts: make(map[string]int),
	}

	// Count principles
	principlesNode := findNode(root, "principles")
	if principlesNode != nil && principlesNode.Kind == yaml.SequenceNode {
		summary.Counts["principles"] = len(principlesNode.Content)

		// Count by priority
		for _, principle := range principlesNode.Content {
			priorityNode := findNode(principle, "priority")
			if priorityNode != nil {
				switch priorityNode.Value {
				case "NON-NEGOTIABLE":
					summary.Counts["non_negotiable_principles"]++
				case "MUST":
					summary.Counts["must_principles"]++
				case "SHOULD":
					summary.Counts["should_principles"]++
				case "MAY":
					summary.Counts["may_principles"]++
				}
			}
		}
	}

	// Count sections
	sectionsNode := findNode(root, "sections")
	if sectionsNode != nil && sectionsNode.Kind == yaml.SequenceNode {
		summary.Counts["sections"] = len(sectionsNode.Content)
	}

	return summary
}
