// Copyright 2026 The OpenKVLab Authors.
//
// Use of this software is governed by the OpenKVLab Software License
// included in the LICENSE file.

package raftpb

// EnsureConfState ensures that cs and all of its pointer fields are non-nil.
// If cs is nil, a new ConfState is allocated. Any nil pointer field is set to
// point to its zero value. Returns the resulting cs.
func EnsureConfState(cs *ConfState) *ConfState {
	if cs == nil {
		cs = new(ConfState)
	}
	if cs.AutoLeave == nil {
		cs.AutoLeave = new(false)
	}
	return cs
}

// EnsureSnapshotMetadata ensures that m and all of its pointer fields are
// non-nil. If m is nil, a new SnapshotMetadata is allocated. Any nil pointer
// field is set to point to its zero value. Returns the resulting m.
func EnsureSnapshotMetadata(m *SnapshotMetadata) *SnapshotMetadata {
	if m == nil {
		m = new(SnapshotMetadata)
	}
	m.ConfState = EnsureConfState(m.ConfState)
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


// EntrySliceToPointers converts a slice of Entry values to a slice of Entry pointers.
// TODO: remove this function
func EntrySliceToPointers(ents []Entry) []*Entry {
	if ents == nil {
		return nil
	}
	result := make([]*Entry, len(ents))
	for i := range ents {
		result[i] = &ents[i]
	}
	return result
}

// EntrySliceFromPointers converts a slice of Entry pointers to a slice of Entry values.
// TODO: remove this function
func EntrySliceFromPointers(ents []*Entry) []Entry {
	if ents == nil {
		return nil
	}
	result := make([]Entry, len(ents))
	for i := range ents {
		if ents[i] != nil {
			result[i] = *ents[i]
		}
	}
	return result
}

// MessageSliceToPointers converts a slice of Message values to a slice of Message pointers.
// TODO: remove this function
func MessageSliceToPointers(msgs []Message) []*Message {
	if msgs == nil {
		return nil
	}
	result := make([]*Message, len(msgs))
	for i := range msgs {
		result[i] = &msgs[i]
	}
	return result
}

// MessageSliceFromPointers converts a slice of Message pointers to a slice of Message values.
// TODO: remove this function
func MessageSliceFromPointers(msgs []*Message) []Message {
	if msgs == nil {
		return nil
	}
	result := make([]Message, len(msgs))
	for i := range msgs {
		if msgs[i] != nil {
			result[i] = *msgs[i]
		}
	}
	return result
}
