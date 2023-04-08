package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap/zapcore"
)

// Adds wrappers for gRPC types to log with zap.

type versionLogger struct {
	Owner     string
	Timestamp int64
	Counter   uint64
}

func newVersionLogger(v *rpc.Version) *versionLogger {
	return &versionLogger{
		Owner:     v.Owner,
		Timestamp: v.Timestamp,
		Counter:   v.Counter,
	}
}

func (l versionLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("owner", l.Owner)
	e.AddInt64("timestamp", l.Timestamp)
	e.AddUint64("counter", l.Counter)
	return nil
}

type metadataLogger map[string]string

func (m metadataLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	for k, v := range m {
		e.AddString(k, v)
	}
	return nil
}

type memberLogger struct {
	ID       string
	Metadata metadataLogger
}

func newMemberLogger(m *rpc.Member) *memberLogger {
	return &memberLogger{
		ID:       m.Id,
		Metadata: m.Metadata,
	}
}

func (l memberLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("id", l.ID)
	if err := e.AddObject("metadata", metadataLogger(l.Metadata)); err != nil {
		return err
	}
	return nil
}

type remoteMemberUpdateLogger struct {
	UpdateType rpc.MemberUpdateType
	Member     *memberLogger
	Version    *versionLogger
}

func newRemoteMemberUpdateLogger(u *rpc.RemoteMemberUpdate) *remoteMemberUpdateLogger {
	return &remoteMemberUpdateLogger{
		UpdateType: u.UpdateType,
		Member:     newMemberLogger(u.Member),
		Version:    newVersionLogger(u.Version),
	}
}

func (l remoteMemberUpdateLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("update-type", l.UpdateType.String())
	if err := e.AddObject("member", l.Member); err != nil {
		return err
	}
	if err := e.AddObject("version", l.Version); err != nil {
		return err
	}
	return nil
}
