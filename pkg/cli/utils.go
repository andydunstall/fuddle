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

package cli

import (
	"net"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getSystemAddress() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	return ln.Addr().String()
}

func loggerWithPath(path string, verbose bool) *zap.Logger {
	loggerConf := zap.NewProductionConfig()
	if startVerbose {
		loggerConf.Level.SetLevel(zapcore.DebugLevel)
	}
	loggerConf.OutputPaths = []string{path}
	return zap.Must(loggerConf.Build())
}