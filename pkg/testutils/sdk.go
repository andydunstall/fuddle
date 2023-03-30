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

package testutils

import (
	"context"

	fuddle "github.com/fuddle-io/fuddle-go"
)

func WaitForMembers(ctx context.Context, client *fuddle.Fuddle, count int) error {
	found := false
	recvCh := make(chan interface{})
	unsubscribe := client.Subscribe(func() {
		if found {
			return
		}

		if len(client.Members()) == count {
			found = true
			close(recvCh)
			return
		}
	})
	defer unsubscribe()

	if err := WaitWithContext(ctx, recvCh); err != nil {
		return err
	}
	return nil
}
