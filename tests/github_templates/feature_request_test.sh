#!/usr/bin/env bats
# Tests for feature_request.md template

# Load validation library
source "$(dirname "$BATS_TEST_DIRNAME")/lib/validation_lib.sh"

FEATURE_REQUEST_PATH=".github/ISSUE_TEMPLATE/feature_request.md"

@test "feature_request.md file exists" {
  [ -f "$FEATURE_REQUEST_PATH" ]
}

@test "feature_request.md has valid YAML frontmatter" {
  run validate_yaml_syntax "$FEATURE_REQUEST_PATH"
  [ "$status" -eq 0 ]
}

@test "feature_request.md has required YAML fields (name, about)" {
  run validate_required_fields "$FEATURE_REQUEST_PATH"
  [ "$status" -eq 0 ]
}

@test "feature_request.md has all required sections" {
  run validate_template_sections "$FEATURE_REQUEST_PATH" feature_request
  [ "$status" -eq 0 ]
}

@test "feature_request.md has Problem Statement section" {
  grep -qF "## Problem Statement" "$FEATURE_REQUEST_PATH"
}

@test "feature_request.md has Use Case section" {
  grep -qF "## Use Case" "$FEATURE_REQUEST_PATH"
}

@test "feature_request.md has Proposed Solution section" {
  grep -qF "## Proposed Solution" "$FEATURE_REQUEST_PATH"
}

@test "feature_request.md has Alternatives Considered section" {
  grep -qF "## Alternatives Considered" "$FEATURE_REQUEST_PATH"
}

@test "feature_request.md has Additional Context section" {
  grep -qF "## Additional Context" "$FEATURE_REQUEST_PATH"
}

@test "feature_request.md YAML has 'name' field" {
  if command -v yq >/dev/null 2>&1; then
    result=$(yq -r '.name' <(sed -n '/^---$/,/^---$/p' "$FEATURE_REQUEST_PATH" | sed '1d;$d'))
    [ "$result" != "null" ]
    [ -n "$result" ]
  else
    skip "yq not available"
  fi
}

@test "feature_request.md YAML has 'about' field" {
  if command -v yq >/dev/null 2>&1; then
    result=$(yq -r '.about' <(sed -n '/^---$/,/^---$/p' "$FEATURE_REQUEST_PATH" | sed '1d;$d'))
    [ "$result" != "null" ]
    [ -n "$result" ]
  else
    skip "yq not available"
  fi
}

@test "feature_request.md YAML has 'title' field (optional)" {
  if command -v yq >/dev/null 2>&1; then
    # This is optional, so we just check it doesn't error
    yq -r '.title' <(sed -n '/^---$/,/^---$/p' "$FEATURE_REQUEST_PATH" | sed '1d;$d') >/dev/null
  else
    skip "yq not available"
  fi
}

@test "feature_request.md YAML has 'labels' field (optional)" {
  if command -v yq >/dev/null 2>&1; then
    # This is optional, so we just check it doesn't error
    yq -r '.labels' <(sed -n '/^---$/,/^---$/p' "$FEATURE_REQUEST_PATH" | sed '1d;$d') >/dev/null
  else
    skip "yq not available"
  fi
}
