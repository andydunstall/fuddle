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
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Logger() *zap.Logger {
	logLevel := os.Getenv("FUDDLE_LOG_LEVEL")

	loggerConf := zap.NewProductionConfig()
	switch logLevel {
	case "debug":
		loggerConf.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		loggerConf.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		loggerConf.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		loggerConf.Level.SetLevel(zapcore.ErrorLevel)
	default:
		// If the level is invalid or not specified don't use a logger.
		return zap.NewNop()
	}
	return zap.Must(loggerConf.Build())
}
