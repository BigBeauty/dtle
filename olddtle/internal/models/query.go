/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/actiontech/dtle, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package models

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	hcodec "github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
	"github.com/ugorji/go/codec"
)

var (
	ErrNoLeader     = fmt.Errorf("No cluster leader")
	ErrNoRegionPath = fmt.Errorf("No path to region")
)

type MessageType uint8

const (
	NodeRegisterRequestType MessageType = iota
	NodeDeregisterRequestType
	NodeUpdateStatusRequestType
	JobUpdateStatusRequestType
	JobRegisterRequestType
	JobDeregisterRequestType
	JobRenewalRequestType
	JobClientUpdateRequestType
	OrderRegisterRequestType
	OrderDeregisterRequestType
	EvalUpdateRequestType
	EvalDeleteRequestType
	AllocUpdateRequestType
	AllocClientUpdateRequestType
)

const (
	// IgnoreUnknownTypeFlag is set along with a MessageType
	// to indicate that the message type can be safely ignored
	// if it is not recognized. This is for future proofing, so
	// that new commands can be added in a way that won't cause
	// old servers to crash when the FSM attempts to process them.
	IgnoreUnknownTypeFlag MessageType = 128
)

// RPCInfo is used to describe common information about query
type RPCInfo interface {
	RequestRegion() string
	IsRead() bool
	AllowStaleRead() bool
}

// QueryOptions is used to specify various flags for read queries
type QueryOptions struct {
	// The target region for this query
	Region string

	// If set, wait until query exceeds given index. Must be provided
	// with MaxQueryTime.
	MinQueryIndex uint64

	// Provided with MinQueryIndex to wait for change.
	MaxQueryTime time.Duration

	// If set, any follower can service the request. Results
	// may be arbitrarily stale.
	AllowStale bool

	// If set, used as prefix for resource list searches
	Prefix string
}

func (q QueryOptions) RequestRegion() string {
	return q.Region
}

// QueryOption only applies to reads, so always true
func (q QueryOptions) IsRead() bool {
	return true
}

func (q QueryOptions) AllowStaleRead() bool {
	return q.AllowStale
}

type WriteRequest struct {
	// The target region for this write
	Region string
}

func (w WriteRequest) RequestRegion() string {
	// The target region for this request
	return w.Region
}

// WriteRequest only applies to writes, always false
func (w WriteRequest) IsRead() bool {
	return false
}

func (w WriteRequest) AllowStaleRead() bool {
	return false
}

// QueryMeta allows a query response to include potentially
// useful metadata about a query
type QueryMeta struct {
	// This is the index associated with the read
	Index uint64

	// If AllowStale is used, this is time elapsed since
	// last contact between the follower and leader. This
	// can be used to gauge staleness.
	LastContact time.Duration

	// Used to indicate if there is a known leader node
	KnownLeader bool
}

// WriteMeta allows a write response to include potentially
// useful metadata about the write
type WriteMeta struct {
	// This is the index associated with the write
	Index uint64
}

// GenericRequest is used to request where no
// specific information is needed.
type GenericRequest struct {
	QueryOptions
}

// GenericResponse is used to respond to a request where no
// specific response information is needed.
type GenericResponse struct {
	WriteMeta
}

// VersionResponse is used for the Status.Version reseponse
type VersionResponse struct {
	Build    string
	Versions map[string]int
	QueryMeta
}

// msgpackHandle is a shared handle for encoding/decoding of structs
var MsgpackHandle = func() *codec.MsgpackHandle {
	h := &codec.MsgpackHandle{RawToString: true}

	// Sets the default type for decoding a map into a nil interface{}.
	// This is necessary in particular because we store the driver configs as a
	// nil interface{}.
	h.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return h
}()

var HashiMsgpackHandle = func() *hcodec.MsgpackHandle {
	h := &hcodec.MsgpackHandle{RawToString: true}

	// Sets the default type for decoding a map into a nil interface{}.
	// This is necessary in particular because we store the driver configs as a
	// nil interface{}.
	h.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return h
}()

// Decode is used to decode a MsgPack encoded object
func Decode(buf []byte, out interface{}) error {
	return codec.NewDecoder(bytes.NewReader(buf), MsgpackHandle).Decode(out)
}

// Encode is used to encode a MsgPack object with type prefix
func Encode(t MessageType, msg interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(uint8(t))
	err := codec.NewEncoder(&buf, MsgpackHandle).Encode(msg)
	return buf.Bytes(), err
}

// RecoverableError wraps an error and marks whether it is recoverable and could
// be retried or it is fatal.
type RecoverableError struct {
	Err         string
	Recoverable bool
}

// NewRecoverableError is used to wrap an error and mark it as recoverable or
// not.
func NewRecoverableError(e error, recoverable bool) error {
	if e == nil {
		return nil
	}

	return &RecoverableError{
		Err:         e.Error(),
		Recoverable: recoverable,
	}
}

// WrapRecoverable wraps an existing error in a new RecoverableError with a new
// message. If the error was recoverable before the returned error is as well;
// otherwise it is unrecoverable.
func WrapRecoverable(msg string, err error) error {
	return &RecoverableError{Err: msg, Recoverable: IsRecoverable(err)}
}

func (r *RecoverableError) Error() string {
	return r.Err
}

func (r *RecoverableError) IsRecoverable() bool {
	return r.Recoverable
}

// Recoverable is an interface for errors to implement to indicate whether or
// not they are fatal or recoverable.
type Recoverable interface {
	error
	IsRecoverable() bool
}

// IsRecoverable returns true if error is a RecoverableError with
// Recoverable=true. Otherwise false is returned.
func IsRecoverable(e error) bool {
	if re, ok := e.(Recoverable); ok {
		return re.IsRecoverable()
	}
	return false
}

// RaftServer has information about a server in the Raft configuration.
type RaftServer struct {
	// ID is the unique ID for the server. These are currently the same
	// as the address, but they will be changed to a real GUID in a future
	// release of Udup.
	ID raft.ServerID

	// Node is the node name of the server, as known by Udup, or this
	// will be set to "(unknown)" otherwise.
	Node string

	// Address is the IP:port of the server, used for Raft communications.
	Address raft.ServerAddress

	// Leader is true if this server is the current cluster leader.
	Leader bool

	// Voter is true if this server has a vote in the cluster. This might
	// be false if the server is staging and still coming online, or if
	// it's a non-voting server, which will be added in a future release of
	// Udup.
	Voter bool
}

// RaftConfigrationResponse is returned when querying for the current Raft
// configuration.
type RaftConfigurationResponse struct {
	// Servers has the list of servers in the Raft configuration.
	Servers []*RaftServer

	// Index has the Raft index of this configuration.
	Index uint64
}

// RaftPeerByAddressRequest is used by the Operator endpoint to apply a Raft
// operation on a specific Raft peer by address in the form of "IP:port".
type RaftPeerByAddressRequest struct {
	// Address is the peer to remove, in the form "IP:port".
	Address raft.ServerAddress

	// WriteRequest holds the Region for this request.
	WriteRequest
}

// RaftRemovePeerRequest is used by the Operator endpoint to apply a Raft
// operation on a specific Raft peer by address in the form of "IP:port".
type RaftRemovePeerRequest struct {
	// Datacenter is the target this request is intended for.
	Datacenter string

	// Address is the peer to remove, in the form "IP:port".
	Address raft.ServerAddress

	// ID is the peer ID to remove.
	ID raft.ServerID

	// WriteRequest holds the ACL token to go along with this request.
	WriteRequest
}
