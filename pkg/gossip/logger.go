package gossip

import (
	"io"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

type loggerWriter struct {
	re     *regexp.Regexp
	logger *zap.Logger
}

func newLoggerWriter(logger *zap.Logger) *loggerWriter {
	re := regexp.MustCompile(`^.*(\[([DEBUG|INFO|WARN|ERR]+)\] )(.*)$`)
	return &loggerWriter{
		re:     re,
		logger: logger,
	}
}

func (w *loggerWriter) Write(b []byte) (n int, err error) {
	s := strings.TrimSpace(string(b))
	matches := w.re.FindStringSubmatch(s)
	if len(matches) < 4 {
		w.logger.Error("unrecognised log", zap.String("log", s))
	}

	switch matches[2] {
	case "DEBUG":
		w.logger.Debug(matches[3])
	case "INFO":
		w.logger.Info(matches[3])
	case "WARN":
		w.logger.Warn(matches[3])
	case "ERR":
		w.logger.Error(matches[3])
	}

	return len(b), nil
}

var _ io.Writer = &loggerWriter{}
