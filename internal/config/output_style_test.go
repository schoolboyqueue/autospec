package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOutputStyle(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"empty string is valid":       {input: "", wantErr: false},
		"default is valid":            {input: "default", wantErr: false},
		"compact is valid":            {input: "compact", wantErr: false},
		"minimal is valid":            {input: "minimal", wantErr: false},
		"plain is valid":              {input: "plain", wantErr: false},
		"raw is valid":                {input: "raw", wantErr: false},
		"uppercase DEFAULT is valid":  {input: "DEFAULT", wantErr: false},
		"mixed case Compact is valid": {input: "Compact", wantErr: false},
		"whitespace trimmed":          {input: "  plain  ", wantErr: false},
		"invalid style errors":        {input: "invalid", wantErr: true},
		"json style errors":           {input: "json", wantErr: true},
		"verbose style errors":        {input: "verbose", wantErr: true},
		"number errors":               {input: "123", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := ValidateOutputStyle(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid output_style")
				assert.Contains(t, err.Error(), "valid options:")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeOutputStyle(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		want    OutputStyle
		wantErr bool
	}{
		"empty returns default":      {input: "", want: OutputStyleDefault, wantErr: false},
		"default normalized":         {input: "default", want: OutputStyleDefault, wantErr: false},
		"compact normalized":         {input: "compact", want: OutputStyleCompact, wantErr: false},
		"minimal normalized":         {input: "minimal", want: OutputStyleMinimal, wantErr: false},
		"plain normalized":           {input: "plain", want: OutputStylePlain, wantErr: false},
		"raw normalized":             {input: "raw", want: OutputStyleRaw, wantErr: false},
		"uppercase normalized":       {input: "PLAIN", want: OutputStylePlain, wantErr: false},
		"mixed case normalized":      {input: "MiNiMaL", want: OutputStyleMinimal, wantErr: false},
		"whitespace trimmed":         {input: "  raw  ", want: OutputStyleRaw, wantErr: false},
		"invalid returns error":      {input: "invalid", want: "", wantErr: true},
		"special chars return error": {input: "def@ult", want: "", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeOutputStyle(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid output_style")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestOutputStyle_IsRaw(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		style OutputStyle
		want  bool
	}{
		"raw is raw":         {style: OutputStyleRaw, want: true},
		"default is not raw": {style: OutputStyleDefault, want: false},
		"compact is not raw": {style: OutputStyleCompact, want: false},
		"minimal is not raw": {style: OutputStyleMinimal, want: false},
		"plain is not raw":   {style: OutputStylePlain, want: false},
		"empty is not raw":   {style: "", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.style.IsRaw())
		})
	}
}

func TestOutputStyle_String(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		style OutputStyle
		want  string
	}{
		"default": {style: OutputStyleDefault, want: "default"},
		"compact": {style: OutputStyleCompact, want: "compact"},
		"minimal": {style: OutputStyleMinimal, want: "minimal"},
		"plain":   {style: OutputStylePlain, want: "plain"},
		"raw":     {style: OutputStyleRaw, want: "raw"},
		"empty":   {style: "", want: ""},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.style.String())
		})
	}
}

func TestValidOutputStyleNames(t *testing.T) {
	t.Parallel()

	names := ValidOutputStyleNames()
	assert.Equal(t, []string{"default", "compact", "minimal", "plain", "raw"}, names)
	assert.Len(t, names, 5)
}
