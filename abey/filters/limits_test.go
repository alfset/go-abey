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
	"math/big"
	"testing"
	"time"

	"github.com/AbeyFoundation/go-abey/abeydb"
	"github.com/AbeyFoundation/go-abey/common"
	"github.com/AbeyFoundation/go-abey/core/bloombits"
	"github.com/AbeyFoundation/go-abey/core/types"
	"github.com/AbeyFoundation/go-abey/event"
	"github.com/AbeyFoundation/go-abey/rpc"
)

// stubBackend is a Backend that only knows about a chain head. The limit checks
// under test run before any block data is touched, so everything else is inert.
type stubBackend struct {
	head int64
	mux  *event.TypeMux

	txsFeed     event.Feed
	logsFeed    event.Feed
	rmLogsFeed  event.Feed
	chainFeed   event.Feed
	headerCalls int
}

func newStubBackend(head int64) *stubBackend {
	return &stubBackend{head: head, mux: new(event.TypeMux)}
}

func (b *stubBackend) ChainDb() abeydb.Database      { return nil }
func (b *stubBackend) EventMux() *event.TypeMux      { return b.mux }
func (b *stubBackend) BloomStatus() (uint64, uint64) { return 4096, 0 }

func (b *stubBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	b.headerCalls++
	n := number.Int64()
	if n < 0 {
		n = b.head
	}
	return &types.Header{Number: big.NewInt(n)}, nil
}

func (b *stubBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return nil, nil
}
func (b *stubBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return nil, nil
}
func (b *stubBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	return nil, nil
}
func (b *stubBackend) SubscribeNewTxsEvent(ch chan<- types.NewTxsEvent) event.Subscription {
	return b.txsFeed.Subscribe(ch)
}
func (b *stubBackend) SubscribeChainEvent(ch chan<- types.FastChainEvent) event.Subscription {
	return b.chainFeed.Subscribe(ch)
}
func (b *stubBackend) SubscribeRemovedLogsEvent(ch chan<- types.RemovedLogsEvent) event.Subscription {
	return b.rmLogsFeed.Subscribe(ch)
}
func (b *stubBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.logsFeed.Subscribe(ch)
}
func (b *stubBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {}

// withLimits installs limits for the duration of a test and restores the
// previous ones afterwards, since they are process-wide.
func withLimits(t *testing.T, l Limits) {
	t.Helper()

	prev := currentLimits()
	SetLimits(l)
	t.Cleanup(func() { SetLimits(prev) })
}

func TestResolveRange(t *testing.T) {
	const head = 10000

	withLimits(t, Limits{MaxBlockRange: 500, Timeout: time.Second})
	backend := newStubBackend(head)

	num := func(n int64) *big.Int { return big.NewInt(n) }
	latest := num(rpc.LatestBlockNumber.Int64())
	pending := num(rpc.PendingBlockNumber.Int64())

	tests := []struct {
		name      string
		from, to  *big.Int
		wantBegin int64
		wantEnd   int64
		wantErr   bool
	}{
		// The regression this whole change exists for: "latest" is the sentinel -1,
		// so an unresolved comparison sees to <= from, skips the range check and
		// scans the entire chain.
		{name: "earliest to latest is rejected", from: num(0), to: latest, wantErr: true},
		{name: "earliest to pending is rejected", from: num(0), to: pending, wantErr: true},
		{name: "earliest to omitted is rejected", from: num(0), to: nil, wantErr: true},

		{name: "narrow range is allowed", from: num(100), to: num(200), wantBegin: 100, wantEnd: 200},
		{name: "range exactly at the cap is allowed", from: num(0), to: num(500), wantBegin: 0, wantEnd: 500},
		{name: "range one past the cap is rejected", from: num(0), to: num(501), wantErr: true},

		// "latest" is resolved, not rejected: the common indexer pattern of polling
		// from the last seen block up to the head keeps working, so long as the
		// resolved span fits the cap.
		{name: "narrow range up to latest is allowed", from: num(head - 100), to: latest, wantBegin: head - 100, wantEnd: head},
		{name: "range up to latest exactly at the cap is allowed", from: num(head - 500), to: latest, wantBegin: head - 500, wantEnd: head},
		{name: "range up to latest one past the cap is rejected", from: num(head - 501), to: latest, wantErr: true},

		{name: "latest to latest is a single block", from: latest, to: latest, wantBegin: head, wantEnd: head},
		{name: "both omitted default to head", from: nil, to: nil, wantBegin: head, wantEnd: head},
		{name: "pending clamps to head", from: pending, to: pending, wantBegin: head, wantEnd: head},

		{name: "bounds past head clamp to head", from: num(head + 5000), to: num(head + 9000), wantBegin: head, wantEnd: head},
		{name: "from after to is rejected", from: num(600), to: num(100), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			begin, end, err := resolveRange(context.Background(), backend, tt.from, tt.to)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got range [%d, %d]", begin, end)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if begin != tt.wantBegin || end != tt.wantEnd {
				t.Fatalf("got range [%d, %d], want [%d, %d]", begin, end, tt.wantBegin, tt.wantEnd)
			}
		})
	}
}

func TestResolveRangeUnlimited(t *testing.T) {
	const head = 10000

	withLimits(t, Limits{MaxBlockRange: 0, Timeout: time.Second})

	begin, end, err := resolveRange(context.Background(), newStubBackend(head), big.NewInt(0), nil)
	if err != nil {
		t.Fatalf("a zero cap should disable the range check, got %v", err)
	}
	if begin != 0 || end != head {
		t.Fatalf("got range [%d, %d], want [0, %d]", begin, end, head)
	}
}

// TestGetLogsRejectsWideRange exercises the guard through the RPC entry point,
// confirming a wide query is refused before any block is scanned.
func TestGetLogsRejectsWideRange(t *testing.T) {
	withLimits(t, Limits{MaxBlockRange: 500, Timeout: time.Second})

	backend := newStubBackend(10000)
	api := NewPublicFilterAPI(backend, false)

	logs, err := api.GetLogs(context.Background(), FilterCriteria{
		FromBlock: big.NewInt(0),
		ToBlock:   big.NewInt(rpc.LatestBlockNumber.Int64()),
	})
	if err == nil {
		t.Fatalf("expected {from: 0, to: latest} to be rejected, got %d logs", len(logs))
	}
	// Resolving the head costs one lookup; anything beyond that means the scan
	// started before the range was validated, which is the expensive path.
	if backend.headerCalls != 1 {
		t.Fatalf("query should be rejected before scanning, saw %d header lookups", backend.headerCalls)
	}
}

func TestRecordMatches(t *testing.T) {
	f := &Filter{maxLogs: 10}

	if err := f.recordMatches(6); err != nil {
		t.Fatalf("unexpected error below the cap: %v", err)
	}
	// The cap applies to the running total across the whole scan, not to a single
	// block's worth of matches.
	if err := f.recordMatches(4); err != nil {
		t.Fatalf("unexpected error exactly at the cap: %v", err)
	}
	err := f.recordMatches(1)
	if err == nil {
		t.Fatal("expected an error once the cap is passed")
	}
	if !errors.Is(err, ErrLogsLimitExceeded) {
		t.Fatalf("error should wrap ErrLogsLimitExceeded, got %v", err)
	}
}

func TestRecordMatchesUnlimited(t *testing.T) {
	f := &Filter{maxLogs: 0}

	if err := f.recordMatches(1 << 20); err != nil {
		t.Fatalf("a zero cap should disable the log limit, got %v", err)
	}
}
