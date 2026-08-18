// Copyright 2024 The go-abey Authors
// This file is part of the go-abey library.
//
// The go-abey library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-abey library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-abey library. If not, see <http://www.gnu.org/licenses/>.

package abeyapi

import (
	"sync"
	"time"
)

// RPCLimits bounds the cost of the RPC methods that execute EVM code on behalf
// of a caller -- abey_call and abey_estimateGas.
//
// These run against a gas pool of MaxUint64 rather than a block's gas limit,
// because the caller is not paying for them. Without a ceiling the caller
// chooses how much work the node performs, which makes a single request enough
// to occupy a core indefinitely.
type RPCLimits struct {
	// GasCap is the highest gas allowance a single call may execute with. It also
	// caps the upper bound of the estimate_gas binary search, whose starting point
	// is otherwise taken straight from the request. Zero disables the cap.
	GasCap uint64

	// EVMTimeout bounds the wall-clock time spent executing EVM code for one
	// request. For estimate_gas it covers the entire binary search rather than a
	// single iteration, so the worst case stays bounded no matter how many rounds
	// the search needs.
	EVMTimeout time.Duration
}

// DefaultRPCLimits mirrors the ceilings upstream go-ethereum applies. The gas cap
// is far above any real call -- several times a full block -- so it only ever
// truncates allowances that were never going to be paid for anyway.
var DefaultRPCLimits = RPCLimits{
	GasCap:     50000000,
	EVMTimeout: 5 * time.Second,
}

var (
	rpcLimitsMu sync.RWMutex
	rpcLimits   = DefaultRPCLimits
)

// SetRPCLimits installs the EVM execution limits. It must be called during node
// setup, before the RPC endpoints begin serving requests.
func SetRPCLimits(l RPCLimits) {
	rpcLimitsMu.Lock()
	defer rpcLimitsMu.Unlock()

	rpcLimits = l
}

// currentRPCLimits returns a copy of the active limits.
func currentRPCLimits() RPCLimits {
	rpcLimitsMu.RLock()
	defer rpcLimitsMu.RUnlock()

	return rpcLimits
}
