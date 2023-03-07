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
	"regexp"
	"strings"
)

func wildcardToRegex(pattern string) string {
	components := strings.Split(pattern, "*")
	if len(components) == 1 {
		return "^" + regexp.QuoteMeta(pattern) + "$"
	}
	var regex strings.Builder
	for i, component := range components {
		// Replace * with .*
		if i > 0 {
			regex.WriteString(".*")
		}
		regex.WriteString(regexp.QuoteMeta(component))
	}
	return "^" + regex.String() + "$"
}

// Match returns true if the wildcard pattern matches the given value, false
// otherwise.
func Match(pattern string, value string) bool {
	match, _ := regexp.MatchString(wildcardToRegex(pattern), value)
	return match
}
