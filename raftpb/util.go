// Copyright 2026 The OpenKVLab Authors.
//
// Use of this software is governed by the OpenKVLab Software License
// included in the LICENSE file.

package raftpb

// EnsureSnapshotMetadata ensures that m and all of its pointer fields are
// non-nil. If m is nil, a new SnapshotMetadata is allocated. Any nil pointer
// field is set to point to its zero value. Returns the resulting m.
func EnsureSnapshotMetadata(m *SnapshotMetadata) *SnapshotMetadata {
	if m == nil {
		m = new(SnapshotMetadata)
	}
	if m.ConfState == nil {
		m.ConfState = new(ConfState)
	}
	if m.Index == nil {
		m.Index = new(uint64)
	}
	if m.Term == nil {
		m.Term = new(uint64)
	}
	return m
}

// EnsureSnapshot ensures that s and all of its pointer fields are non-nil.
// If s is nil, a new Snapshot is allocated. Any nil pointer field is set to
// point to its zero value. Returns the resulting s.
func EnsureSnapshot(s *Snapshot) *Snapshot {
	if s == nil {
		s = new(Snapshot)
	}
	s.Metadata = EnsureSnapshotMetadata(s.Metadata)
	return s
}
