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

package filters

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/AbeyFoundation/go-abey/rpc"
)

// Log filtering is the most expensive read path the node exposes: every candidate
// block costs a bloom lookup, a receipt read from disk and an RLP decode, and the
// matched logs are buffered in memory until the whole range has been scanned. A
// burst of wide-range queries can therefore exhaust CPU, memory and disk IOPS and
// starve every other RPC user. Limits bounds what a single query may consume and
// how many such queries may run at once.
type Limits struct {
	// MaxBlockRange is the widest block span a single query may cover. Zero
	// disables the check.
	MaxBlockRange int64

	// MaxLogs is the largest number of matching logs a single query may collect.
	// It is enforced during the scan rather than on the finished result, so an
	// over-wide query is aborted before it can balloon node memory. Exceeding it
	// fails the query outright; a partial result is never returned. Zero disables
	// the check.
	//
	// This is a safety fuse, not a usage quota -- the block range cap is what
	// bounds ordinary queries. The range cap alone cannot bound memory, because it
	// limits how many blocks are scanned but not how many logs each one holds: at
	// the 375 gas floor of a LOG opcode a single block can carry tens of thousands
	// of them. Size this above the busiest legitimate query so it only ever fires
	// on the pathological case.
	MaxLogs int

	// Timeout bounds the wall-clock time a single query may run for. It cancels
	// the scan itself, unlike the HTTP server write timeout which only closes the
	// connection and leaves the scan burning resources in the background.
	Timeout time.Duration
}

// DefaultLimits are tuned for a publicly reachable RPC node, and sized so that
// ordinary indexer traffic passes untouched: a fully saturated block carries on
// the order of 200 logs, so an unfiltered scan of the widest allowed range lands
// around 100k. Operators serving only trusted internal traffic can raise them.
var DefaultLimits = Limits{
	MaxBlockRange: 500,
	MaxLogs:       100000,
	Timeout:       30 * time.Second,
}

var (
	limitsMu sync.RWMutex
	limits   = DefaultLimits
)

// Errors returned to the caller when a query exceeds the configured limits. They
// are deliberately explicit about the active limit so integrators can adjust
// their pagination without guesswork.
var (
	// ErrLogsLimitExceeded is the sentinel wrapped by the error a scan returns once
	// it collects more logs than allowed.
	ErrLogsLimitExceeded = errors.New("log limit exceeded")

	errQueryTimeout  = errors.New("log query timed out, narrow the block range or add filters")
	errHeadUnavailab = errors.New("cannot resolve chain head")
)

// SetLimits installs the log filtering limits. It must be called during node
// setup, before the RPC endpoints begin serving requests.
func SetLimits(l Limits) {
	limitsMu.Lock()
	defer limitsMu.Unlock()

	limits = l
}

// currentLimits returns a copy of the active limits.
func currentLimits() Limits {
	limitsMu.RLock()
	defer limitsMu.RUnlock()

	return limits
}

// resolveRange converts the from/to criteria into concrete block numbers and
// enforces the configured range cap.
//
// Resolving before comparing is the whole point: "latest" and "pending" are
// encoded as the sentinels -1 and -2, so comparing the raw values lets a query
// like {fromBlock: "earliest", toBlock: "latest"} read as (0, -1). A naive
// `to-from > max` check sees to <= from, waves it through, and the filter then
// scans the entire chain.
func resolveRange(ctx context.Context, backend Backend, from, to *big.Int) (int64, int64, error) {
	header, err := backend.HeaderByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		return 0, 0, err
	}
	if header == nil {
		return 0, 0, errHeadUnavailab
	}
	head := header.Number.Int64()

	// Negative values are the latest/pending sentinels, and a number past the head
	// cannot hold logs yet; both clamp to the current head.
	resolve := func(v *big.Int) int64 {
		if v == nil {
			return head
		}
		n := v.Int64()
		if n < 0 || n > head {
			return head
		}
		return n
	}
	begin, end := resolve(from), resolve(to)

	if begin > end {
		return 0, 0, fmt.Errorf("invalid block range: fromBlock (%d) is after toBlock (%d)", begin, end)
	}
	if max := currentLimits().MaxBlockRange; max > 0 && end-begin > max {
		return 0, 0, fmt.Errorf("block range too wide: %d blocks requested, limit is %d", end-begin, max)
	}
	return begin, end, nil
}

// translateQueryErr replaces the opaque context error surfaced by a cancelled
// scan with one that tells the caller what to do about it.
func translateQueryErr(err error) error {
	if err == context.DeadlineExceeded {
		return errQueryTimeout
	}
	return err
}
