package dbutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortFilesNumerically(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "single digit and double digit",
			input: []string{"migrations/10_foo.sql", "migrations/1_bar.sql", "migrations/2_baz.sql", "migrations/9_qux.sql"},
			want:  []string{"migrations/1_bar.sql", "migrations/2_baz.sql", "migrations/9_qux.sql", "migrations/10_foo.sql"},
		},
		{
			name:  "already sorted",
			input: []string{"01_a.sql", "02_b.sql", "03_c.sql"},
			want:  []string{"01_a.sql", "02_b.sql", "03_c.sql"},
		},
		{
			name:  "no numeric prefix goes last",
			input: []string{"readme.md", "1_init.sql", "2_data.sql"},
			want:  []string{"1_init.sql", "2_data.sql", "readme.md"},
		},
		{
			name:  "same number sorts lexicographically",
			input: []string{"1_b.sql", "1_a.sql"},
			want:  []string{"1_a.sql", "1_b.sql"},
		},
		{
			name:  "with directory prefix",
			input: []string{"../../migrations/11_scaling.sql", "../../migrations/2_filter.sql", "../../migrations/1_schema.sql"},
			want:  []string{"../../migrations/1_schema.sql", "../../migrations/2_filter.sql", "../../migrations/11_scaling.sql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := make([]string, len(tt.input))
			copy(files, tt.input)
			sortFilesNumerically(files)
			assert.Equal(t, tt.want, files)
		})
	}
}
