// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package wildcard

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWildcard(t *testing.T) {
	tests := []struct {
		Pattern string
		Value   string
		Match   bool
	}{
		{
			"*", "foo", true,
		},
		{
			"foo.*.car", "foo.bar.car", true,
		},
		{
			"foo.*", "foo.bar.car", true,
		},
		{
			"*.car", "foo.bar.car", true,
		},
		{
			"bar*", "foo", false,
		},
		{
			"foo", "foo", true,
		},
		{
			"(foo)", "foo", false,
		},
		{
			"[*]", "[foo]", true,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("match(%s,%s)", tt.Pattern, tt.Value)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.Match, Match(tt.Pattern, tt.Value))
		})
	}
}
