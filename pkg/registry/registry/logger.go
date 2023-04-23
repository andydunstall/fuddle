package registry

import (
	"strings"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap/zapcore"
)

type memberLogger struct {
	member *rpc.Member2
}

func newMemberLogger(m *rpc.Member2) memberLogger {
	return memberLogger{
		member: m,
	}
}

func (l memberLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("state.id", l.member.State.Id)
	e.AddString("state.status", l.member.State.Status)
	e.AddString("state.service", l.member.State.Service)
	if l.member.State.Locality != nil {
		e.AddString("state.locality.region", l.member.State.Locality.Region)
		e.AddString("state.locality.az", l.member.State.Locality.AvailabilityZone)
	}
	e.AddString("state.started", l.member.State.Service)
	e.AddString("state.revision", l.member.State.Revision)

	e.AddString("liveness", strings.ToLower(l.member.Liveness.String()))

	e.AddString("version.owner", l.member.Version.OwnerId)
	e.AddInt64("version.timestamp", l.member.Version.Timestamp.Timestamp)
	e.AddUint64("version.counter", l.member.Version.Timestamp.Counter)

	e.AddInt64("expiry", l.member.Expiry)

	return nil
}

type memberStateLogger struct {
	state *rpc.MemberState
}

func newMemberStateLogger(s *rpc.MemberState) memberStateLogger {
	return memberStateLogger{
		state: s,
	}
}

func (l memberStateLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("id", l.state.Id)
	e.AddString("status", l.state.Status)
	e.AddString("service", l.state.Service)
	e.AddString("locality.region", l.state.Locality.Region)
	e.AddString("locality.az", l.state.Locality.AvailabilityZone)
	e.AddString("started", l.state.Service)
	e.AddString("revision", l.state.Revision)

	return nil
}

type versionLogger struct {
	version *rpc.Version2
}

func newVersionLogger(v *rpc.Version2) versionLogger {
	return versionLogger{
		version: v,
	}
}

func (l versionLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("owner", l.version.OwnerId)
	e.AddInt64("timestamp", l.version.Timestamp.Timestamp)
	e.AddUint64("counter", l.version.Timestamp.Counter)
	return nil
}
