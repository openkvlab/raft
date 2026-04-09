// This code has been modified from its original form by The OpenKVLab Authors.
// All modifications are Copyright 2026 The OpenKVLab Authors.
//
// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	pb "github.com/openkvlab/raft/raftpb"
)

var testFormatter EntryFormatter = func(data []byte) string {
	return strings.ToUpper(string(data))
}

func TestDescribeEntry(t *testing.T) {
	entry := &pb.Entry{
		Term:  new(uint64(1)),
		Index: new(uint64(2)),
		Type:  pb.EntryType_EntryNormal.Enum(),
		Data:  []byte("hello\x00world"),
	}
	require.Equal(t, `1/2 EntryNormal "hello\x00world"`, DescribeEntry(entry, nil))
	require.Equal(t, "1/2 EntryNormal HELLO\x00WORLD", DescribeEntry(entry, testFormatter))
}

func TestLimitSize(t *testing.T) {
	ents := []*pb.Entry{{Index: new(uint64(4)), Term: new(uint64(4))}, {Index: new(uint64(5)), Term: new(uint64(5))}, {Index: new(uint64(6)), Term: new(uint64(6))}}
	prefix := func(size int) []*pb.Entry {
		return append([]*pb.Entry{}, ents[:size]...) // protect the original slice
	}
	for _, tt := range []struct {
		maxSize uint64
		want    []*pb.Entry
	}{
		{math.MaxUint64, prefix(len(ents))}, // all entries are returned
		// Even if maxSize is zero, the first entry should be returned.
		{0, prefix(1)},
		// Limit to 2.
		{uint64(proto.Size(ents[0]) + proto.Size(ents[1])), prefix(2)},
		{uint64(proto.Size(ents[0]) + proto.Size(ents[1]) + proto.Size(ents[2])/2), prefix(2)},
		{uint64(proto.Size(ents[0]) + proto.Size(ents[1]) + proto.Size(ents[2]) - 1), prefix(2)},
		// All.
		{uint64(proto.Size(ents[0]) + proto.Size(ents[1]) + proto.Size(ents[2])), prefix(3)},
	} {
		t.Run("", func(t *testing.T) {
			got := limitSize(ents, entryEncodingSize(tt.maxSize))
			require.Equal(t, tt.want, got)
			size := entsSize(got)
			require.True(t, len(got) == 1 || size <= entryEncodingSize(tt.maxSize))
		})
	}
}

func TestIsLocalMsg(t *testing.T) {
	tests := []struct {
		msgt    pb.MessageType
		isLocal bool
	}{
		{pb.MessageType_MsgHup, true},
		{pb.MessageType_MsgBeat, true},
		{pb.MessageType_MsgUnreachable, true},
		{pb.MessageType_MsgSnapStatus, true},
		{pb.MessageType_MsgCheckQuorum, true},
		{pb.MessageType_MsgTransferLeader, false},
		{pb.MessageType_MsgProp, false},
		{pb.MessageType_MsgApp, false},
		{pb.MessageType_MsgAppResp, false},
		{pb.MessageType_MsgVote, false},
		{pb.MessageType_MsgVoteResp, false},
		{pb.MessageType_MsgSnap, false},
		{pb.MessageType_MsgHeartbeat, false},
		{pb.MessageType_MsgHeartbeatResp, false},
		{pb.MessageType_MsgTimeoutNow, false},
		{pb.MessageType_MsgReadIndex, false},
		{pb.MessageType_MsgReadIndexResp, false},
		{pb.MessageType_MsgPreVote, false},
		{pb.MessageType_MsgPreVoteResp, false},
		{pb.MessageType_MsgStorageAppend, true},
		{pb.MessageType_MsgStorageAppendResp, true},
		{pb.MessageType_MsgStorageApply, true},
		{pb.MessageType_MsgStorageApplyResp, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.msgt), func(t *testing.T) {
			require.Equal(t, tt.isLocal, IsLocalMsg(tt.msgt))
		})
	}
}

func TestIsResponseMsg(t *testing.T) {
	tests := []struct {
		msgt       pb.MessageType
		isResponse bool
	}{
		{pb.MessageType_MsgHup, false},
		{pb.MessageType_MsgBeat, false},
		{pb.MessageType_MsgUnreachable, true},
		{pb.MessageType_MsgSnapStatus, false},
		{pb.MessageType_MsgCheckQuorum, false},
		{pb.MessageType_MsgTransferLeader, false},
		{pb.MessageType_MsgProp, false},
		{pb.MessageType_MsgApp, false},
		{pb.MessageType_MsgAppResp, true},
		{pb.MessageType_MsgVote, false},
		{pb.MessageType_MsgVoteResp, true},
		{pb.MessageType_MsgSnap, false},
		{pb.MessageType_MsgHeartbeat, false},
		{pb.MessageType_MsgHeartbeatResp, true},
		{pb.MessageType_MsgTimeoutNow, false},
		{pb.MessageType_MsgReadIndex, false},
		{pb.MessageType_MsgReadIndexResp, true},
		{pb.MessageType_MsgPreVote, false},
		{pb.MessageType_MsgPreVoteResp, true},
		{pb.MessageType_MsgStorageAppend, false},
		{pb.MessageType_MsgStorageAppendResp, true},
		{pb.MessageType_MsgStorageApply, false},
		{pb.MessageType_MsgStorageApplyResp, true},
	}

	for i, tt := range tests {
		got := IsResponseMsg(tt.msgt)
		assert.Equal(t, tt.isResponse, got, "#%d", i)
	}
}

// requireProtoEqual checks that two proto.Message values are equal using
// proto.Equal to avoid failures from internal protobuf lazy-initialization state.
func requireProtoEqual(t *testing.T, expected, actual proto.Message, msgAndArgs ...interface{}) {
	t.Helper()
	if !proto.Equal(expected, actual) {
		require.Fail(t, fmt.Sprintf("proto not equal:\nexpected: %v\nactual:   %v", expected, actual), msgAndArgs...)
	}
}

// assertEntriesEqual compares two []*pb.Entry slices using proto.Equal to
// avoid failures from internal protobuf lazy-initialization state.
func assertEntriesEqual(t *testing.T, expected, actual []*pb.Entry, msgAndArgs ...interface{}) {
	t.Helper()
	if len(expected) != len(actual) {
		assert.Fail(t, fmt.Sprintf("entry slice lengths differ: expected %d, got %d", len(expected), len(actual)), msgAndArgs...)
		return
	}
	for i := range expected {
		if !proto.Equal(expected[i], actual[i]) {
			assert.Fail(t, fmt.Sprintf("entry[%d] not equal:\nexpected: %v\nactual:   %v", i, expected[i], actual[i]), msgAndArgs...)
		}
	}
}

// requireEntriesEqual compares two []*pb.Entry slices using proto.Equal to
// avoid failures from internal protobuf lazy-initialization state.
func requireEntriesEqual(t *testing.T, expected, actual []*pb.Entry, msgAndArgs ...interface{}) {
	t.Helper()
	if len(expected) != len(actual) {
		require.Fail(t, fmt.Sprintf("entry slice lengths differ: expected %d, got %d", len(expected), len(actual)), msgAndArgs...)
		return
	}
	for i := range expected {
		if !proto.Equal(expected[i], actual[i]) {
			require.Fail(t, fmt.Sprintf("entry[%d] not equal:\nexpected: %v\nactual:   %v", i, expected[i], actual[i]), msgAndArgs...)
		}
	}
}

// requireReadyEqual compares two Ready structs using proto.Equal for protobuf
// fields to avoid failures from internal protobuf lazy-initialization state.
func requireReadyEqual(t *testing.T, expected, actual Ready, msgAndArgs ...interface{}) {
	t.Helper()
	if !proto.Equal(expected.HardState, actual.HardState) {
		require.Fail(t, fmt.Sprintf("HardState not equal:\nexpected: %v\nactual:   %v", expected.HardState, actual.HardState), msgAndArgs...)
	}
	if len(expected.Entries) != len(actual.Entries) {
		require.Fail(t, fmt.Sprintf("Entries length not equal: expected %d, got %d", len(expected.Entries), len(actual.Entries)), msgAndArgs...)
	} else {
		for i := range expected.Entries {
			if !proto.Equal(expected.Entries[i], actual.Entries[i]) {
				require.Fail(t, fmt.Sprintf("Entries[%d] not equal:\nexpected: %v\nactual:   %v", i, expected.Entries[i], actual.Entries[i]), msgAndArgs...)
			}
		}
	}
	if len(expected.CommittedEntries) != len(actual.CommittedEntries) {
		require.Fail(t, fmt.Sprintf("CommittedEntries length not equal: expected %d, got %d", len(expected.CommittedEntries), len(actual.CommittedEntries)), msgAndArgs...)
	} else {
		for i := range expected.CommittedEntries {
			if !proto.Equal(expected.CommittedEntries[i], actual.CommittedEntries[i]) {
				require.Fail(t, fmt.Sprintf("CommittedEntries[%d] not equal:\nexpected: %v\nactual:   %v", i, expected.CommittedEntries[i], actual.CommittedEntries[i]), msgAndArgs...)
			}
		}
	}
	require.Equal(t, expected.SoftState, actual.SoftState, msgAndArgs...)
	require.Equal(t, expected.ReadStates, actual.ReadStates, msgAndArgs...)
	require.Equal(t, expected.MustSync, actual.MustSync, msgAndArgs...)
}

// TestPayloadSizeOfEmptyEntry ensures that payloadSize of empty entry is always zero.
// This property is important because new leaders append an empty entry to their log,
// and we don't want this to count towards the uncommitted log quota.
func TestPayloadSizeOfEmptyEntry(t *testing.T) {
	e := &pb.Entry{Data: nil}
	require.Equal(t, 0, int(payloadSize(e)))
}
