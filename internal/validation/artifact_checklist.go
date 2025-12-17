package validation

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ChecklistValidator validates checklist.yaml artifacts.
type ChecklistValidator struct {
	baseValidator
}

// Type returns the artifact type.
func (v *ChecklistValidator) Type() ArtifactType {
	return ArtifactTypeChecklist
}

// Validate validates a checklist.yaml file at the given path.
func (v *ChecklistValidator) Validate(path string) *ValidationResult {
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
			Hint:    "The checklist.yaml file should start with key-value pairs, not a list or scalar",
		})
		return result
	}

	// Validate required fields
	checklistNode := validateRequiredField(rootMapping, "checklist", result)
	categoriesNode := validateRequiredField(rootMapping, "categories", result)

	// Validate checklist section
	if checklistNode != nil {
		v.validateChecklistSection(checklistNode, result)
	}

	// Validate categories section
	if categoriesNode != nil {
		v.validateCategories(categoriesNode, result)
	}

	// Build summary if valid
	if result.Valid {
		result.Summary = v.buildSummary(rootMapping)
	}

	return result
}

// validateChecklistSection validates the checklist section.
func (v *ChecklistValidator) validateChecklistSection(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "checklist", yaml.MappingNode, "object", result) {
		return
	}

	// Required fields in checklist
	validateRequiredField(node, "feature", result)
	validateRequiredField(node, "branch", result)
	validateRequiredField(node, "domain", result)

	// Validate audience enum if present
	audienceNode := findNode(node, "audience")
	if audienceNode != nil {
		validateEnumValue(audienceNode, "checklist.audience",
			[]string{"author", "reviewer", "qa", "release"}, result)
	}

	// Validate depth enum if present
	depthNode := findNode(node, "depth")
	if depthNode != nil {
		validateEnumValue(depthNode, "checklist.depth",
			[]string{"lightweight", "standard", "comprehensive"}, result)
	}
}

// validateCategories validates the categories section.
func (v *ChecklistValidator) validateCategories(node *yaml.Node, result *ValidationResult) {
	if !validateFieldType(node, "categories", yaml.SequenceNode, "array", result) {
		return
	}

	for i, categoryNode := range node.Content {
		path := fmt.Sprintf("categories[%d]", i)
		v.validateCategory(categoryNode, path, result)
	}
}

// validateCategory validates a single category.
func (v *ChecklistValidator) validateCategory(node *yaml.Node, path string, result *ValidationResult) {
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
	nameNode := findNode(node, "name")
	if nameNode == nil {
		result.AddError(&ValidationError{
			Path:    fmt.Sprintf("%s.name", path),
			Line:    getNodeLine(node),
			Message: "missing required field: name",
			Hint:    "Add the 'name' field to this category",
		})
	}

	itemsNode := findNode(node, "items")
	if itemsNode == nil {
		result.AddError(&ValidationError{
			Path:    fmt.Sprintf("%s.items", path),
			Line:    getNodeLine(node),
			Message: "missing required field: items",
			Hint:    "Add the 'items' field with a list of checklist items",
		})
	} else if itemsNode.Kind == yaml.SequenceNode {
		for j, itemNode := range itemsNode.Content {
			itemPath := fmt.Sprintf("%s.items[%d]", path, j)
			v.validateChecklistItem(itemNode, itemPath, result)
		}
	}
}

// validateChecklistItem validates a single checklist item.
func (v *ChecklistValidator) validateChecklistItem(node *yaml.Node, path string, result *ValidationResult) {
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
	requiredFields := []string{"id", "description", "status"}
	for _, field := range requiredFields {
		fieldNode := findNode(node, field)
		if fieldNode == nil {
			result.AddError(&ValidationError{
				Path:    fmt.Sprintf("%s.%s", path, field),
				Line:    getNodeLine(node),
				Message: fmt.Sprintf("missing required field: %s", field),
				Hint:    fmt.Sprintf("Add the '%s' field to this checklist item", field),
			})
		}
	}

	// Validate status enum
	statusNode := findNode(node, "status")
	if statusNode != nil {
		validateEnumValue(statusNode, path+".status",
			[]string{"pending", "pass", "fail"}, result)
	}

	// Validate quality_dimension enum if present
	qualityNode := findNode(node, "quality_dimension")
	if qualityNode != nil {
		validateEnumValue(qualityNode, path+".quality_dimension",
			[]string{"completeness", "clarity", "consistency", "measurability", "coverage", "edge_cases"}, result)
	}
}

// buildSummary builds the summary for a valid checklist artifact.
func (v *ChecklistValidator) buildSummary(root *yaml.Node) *ArtifactSummary {
	summary := &ArtifactSummary{
		Type:   ArtifactTypeChecklist,
		Counts: make(map[string]int),
	}

	// Count categories and items
	categoriesNode := findNode(root, "categories")
	if categoriesNode != nil && categoriesNode.Kind == yaml.SequenceNode {
		summary.Counts["categories"] = len(categoriesNode.Content)

		totalItems := 0
		passedItems := 0
		failedItems := 0
		pendingItems := 0

		for _, category := range categoriesNode.Content {
			itemsNode := findNode(category, "items")
			if itemsNode != nil && itemsNode.Kind == yaml.SequenceNode {
				totalItems += len(itemsNode.Content)

				for _, item := range itemsNode.Content {
					statusNode := findNode(item, "status")
					if statusNode != nil {
						switch statusNode.Value {
						case "pass":
							passedItems++
						case "fail":
							failedItems++
						case "pending":
							pendingItems++
						}
					}
				}
			}
		}

		summary.Counts["total_items"] = totalItems
		summary.Counts["passed"] = passedItems
		summary.Counts["failed"] = failedItems
		summary.Counts["pending"] = pendingItems
	}

	return summary
}
