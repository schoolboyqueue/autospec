package validation

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// AnalysisValidator validates analysis.yaml artifacts.
type AnalysisValidator struct {
	baseValidator
}

// Type returns the artifact type.
func (v *AnalysisValidator) Type() ArtifactType {
	return ArtifactTypeAnalysis
}

// Validate validates an analysis.yaml file at the given path.
func (v *AnalysisValidator) Validate(path string) *ValidationResult {
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
			Hint:    "The analysis.yaml file should start with key-value pairs, not a list or scalar",
		})
		return result
	}

	// Validate required fields
	analysisNode := validateRequiredField(rootMapping, "analysis", result)
	findingsNode := validateRequiredField(rootMapping, "findings", result)
	summaryNode := validateRequiredField(rootMapping, "summary", result)

	// Validate analysis section
	if analysisNode != nil {
		v.validateAnalysisSection(analysisNode, result)
	}

	// Validate findings section
	if findingsNode != nil {
		v.validateFindings(findingsNode, result)
	}

	// Validate summary section
	if summaryNode != nil {
		v.validateSummary(summaryNode, result)
	}

	// Build summary if valid
	if result.Valid {
		result.Summary = v.buildSummary(rootMapping)
	}

	return result
}

// validateAnalysisSection validates the analysis section.
func (v *AnalysisValidator) validateAnalysisSection(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "analysis", yaml.MappingNode, "object", result) {
		return
	}

	// Required fields in analysis
	validateRequiredField(node, "branch", result)
	validateRequiredField(node, "timestamp", result)
}

// validateFindings validates the findings section.
func (v *AnalysisValidator) validateFindings(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "findings", yaml.SequenceNode, "array", result) {
		return
	}

	for i, findingNode := range node.Content {
		path := fmt.Sprintf("findings[%d]", i)
		v.validateFinding(findingNode, path, result)
	}
}

// validateFinding validates a single finding.
func (v *AnalysisValidator) validateFinding(node *yaml.Node, path string, result *ValidationResult) {
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
	requiredFields := []string{"id", "category", "severity", "location", "summary"}
	for _, field := range requiredFields {
		fieldNode := findNode(node, field)
		if fieldNode == nil {
			result.AddError(&ValidationError{
				Path:    fmt.Sprintf("%s.%s", path, field),
				Line:    getNodeLine(node),
				Message: fmt.Sprintf("missing required field: %s", field),
				Hint:    fmt.Sprintf("Add the '%s' field to this finding", field),
			})
		}
	}

	// Validate category enum
	categoryNode := findNode(node, "category")
	if categoryNode != nil {
		validateEnumValue(categoryNode, path+".category",
			[]string{"duplication", "ambiguity", "coverage", "constitution", "inconsistency", "underspecification"}, result)
	}

	// Validate severity enum
	severityNode := findNode(node, "severity")
	if severityNode != nil {
		validateEnumValue(severityNode, path+".severity",
			[]string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}, result)
	}
}

// validateSummary validates the summary section.
func (v *AnalysisValidator) validateSummary(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "summary", yaml.MappingNode, "object", result) {
		return
	}

	// Required fields in summary
	statusNode := findNode(node, "overall_status")
	if statusNode == nil {
		result.AddError(&ValidationError{
			Path:    "summary.overall_status",
			Line:    getNodeLine(node),
			Message: "missing required field: overall_status",
			Hint:    "Add the 'overall_status' field with value PASS, WARN, or FAIL",
		})
	} else {
		validateEnumValue(statusNode, "summary.overall_status", []string{"PASS", "WARN", "FAIL"}, result)
	}
}

// buildSummary builds the summary for a valid analysis artifact.
func (v *AnalysisValidator) buildSummary(root *yaml.Node) *ArtifactSummary {
	summary := &ArtifactSummary{
		Type:   ArtifactTypeAnalysis,
		Counts: make(map[string]int),
	}

	// Count findings
	findingsNode := findNode(root, "findings")
	if findingsNode != nil && findingsNode.Kind == yaml.SequenceNode {
		summary.Counts["findings"] = len(findingsNode.Content)

		// Count by severity
		for _, finding := range findingsNode.Content {
			severityNode := findNode(finding, "severity")
			if severityNode != nil {
				switch severityNode.Value {
				case "CRITICAL":
					summary.Counts["critical_findings"]++
				case "HIGH":
					summary.Counts["high_findings"]++
				case "MEDIUM":
					summary.Counts["medium_findings"]++
				case "LOW":
					summary.Counts["low_findings"]++
				}
			}
		}
	}

	return summary
}
