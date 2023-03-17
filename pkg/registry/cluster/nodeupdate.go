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

package cluster

import (
	"go.uber.org/zap/zapcore"
)

type UpdateType string

const (
	UpdateTypeRegister   UpdateType = "register"
	UpdateTypeUnregister UpdateType = "unregister"
	UpdateTypeMetadata   UpdateType = "metadata"
)

type NodeAttributes struct {
	// Service is the type of service running on the node.
	Service string `json:"service,omitempty"`

	// Locality is the location of the node in the cluster.
	Locality string `json:"locality,omitempty"`

	// Created is the time the node was created in UNIX milliseconds.
	Created int64 `json:"created,omitempty"`

	// Revision identifies the version of the service running on the node.
	Revision string `json:"revision,omitempty"`
}

func (a *NodeAttributes) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("service", a.Service)
	e.AddString("locality", a.Locality)
	e.AddInt64("created", a.Created)
	e.AddString("revision", a.Revision)
	return nil
}

type NodeUpdate struct {
	// ID is the ID of the node in the update.
	ID string `json:"id,omitempty"`

	// UpdateType indicates the type of update, either register, unregister or
	// metadata.
	UpdateType UpdateType `json:"update_type,omitempty"`

	// Attributes contains the set of immutable attributes for the node. This
	// will only be included in register updates.
	Attributes *NodeAttributes `json:"attributes,omitempty"`

	// Metadata contains application defined metadata. If the update is type
	// metadata, the field only contains the fields that have been updated.
	Metadata Metadata `json:"metadata,omitempty"`
}

func (u *NodeUpdate) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("id", u.ID)
	e.AddString("update-type", string(u.UpdateType))
	if err := e.AddObject("attributes", u.Attributes); err != nil {
		return err
	}
	if err := e.AddObject("metadata", u.Metadata); err != nil {
		return err
	}
	return nil
}
