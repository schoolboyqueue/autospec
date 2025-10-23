#!/usr/bin/env bats
# Tests for bug_report.md template

# Load validation library
source "$(dirname "$BATS_TEST_DIRNAME")/lib/validation_lib.sh"

BUG_REPORT_PATH=".github/ISSUE_TEMPLATE/bug_report.md"

@test "bug_report.md file exists" {
  [ -f "$BUG_REPORT_PATH" ]
}

@test "bug_report.md has valid YAML frontmatter" {
  run validate_yaml_syntax "$BUG_REPORT_PATH"
  [ "$status" -eq 0 ]
}

@test "bug_report.md has required YAML fields (name, about)" {
  run validate_required_fields "$BUG_REPORT_PATH"
  [ "$status" -eq 0 ]
}

@test "bug_report.md has all required sections" {
  run validate_template_sections "$BUG_REPORT_PATH" bug_report
  [ "$status" -eq 0 ]
}

@test "bug_report.md has Bug Description section" {
  grep -qF "## Bug Description" "$BUG_REPORT_PATH"
}

@test "bug_report.md has Steps to Reproduce section" {
  grep -qF "## Steps to Reproduce" "$BUG_REPORT_PATH"
}

@test "bug_report.md has Expected Behavior section" {
  grep -qF "## Expected Behavior" "$BUG_REPORT_PATH"
}

@test "bug_report.md has Actual Behavior section" {
  grep -qF "## Actual Behavior" "$BUG_REPORT_PATH"
}

@test "bug_report.md has Environment section" {
  grep -qF "## Environment" "$BUG_REPORT_PATH"
}

@test "bug_report.md has Additional Context section" {
  grep -qF "## Additional Context" "$BUG_REPORT_PATH"
}

@test "bug_report.md YAML has 'name' field" {
  if command -v yq >/dev/null 2>&1; then
    result=$(yq eval '.name' "$BUG_REPORT_PATH")
    [ "$result" != "null" ]
  else
    skip "yq not available"
  fi
}

@test "bug_report.md YAML has 'about' field" {
  if command -v yq >/dev/null 2>&1; then
    result=$(yq eval '.about' "$BUG_REPORT_PATH")
    [ "$result" != "null" ]
  else
    skip "yq not available"
  fi
}

@test "bug_report.md YAML has 'title' field (optional)" {
  if command -v yq >/dev/null 2>&1; then
    # This is optional, so we just check it doesn't error
    yq eval '.title' "$BUG_REPORT_PATH" >/dev/null
  else
    skip "yq not available"
  fi
}

@test "bug_report.md YAML has 'labels' field (optional)" {
  if command -v yq >/dev/null 2>&1; then
    # This is optional, so we just check it doesn't error
    yq eval '.labels' "$BUG_REPORT_PATH" >/dev/null
  else
    skip "yq not available"
  fi
}
