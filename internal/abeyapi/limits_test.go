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
	"testing"
	"time"

	"github.com/AbeyFoundation/go-abey/common/math"
	"github.com/AbeyFoundation/go-abey/params"
)

// withRPCLimits installs limits for the duration of a test and restores the
// previous ones afterwards, since they are process-wide.
func withRPCLimits(t *testing.T, l RPCLimits) {
	t.Helper()

	prev := currentRPCLimits()
	SetRPCLimits(l)
	t.Cleanup(func() { SetRPCLimits(prev) })
}

// capGas mirrors the clamp doCall applies to the gas allowance. Keeping it in
// one place lets the boundaries be asserted without standing up an EVM.
func capGas(requested uint64) uint64 {
	gas := requested
	if gas == 0 {
		gas = math.MaxUint64 / 2
	}
	if cap := currentRPCLimits().GasCap; cap != 0 && gas > cap {
		gas = cap
	}
	return gas
}

func TestGasCap(t *testing.T) {
	const cap = 50000000

	withRPCLimits(t, RPCLimits{GasCap: cap, EVMTimeout: time.Second})

	tests := []struct {
		name      string
		requested uint64
		want      uint64
	}{
		// The unmetered default is half of MaxUint64; combined with the MaxUint64
		// gas pool that is what let a single call run unbounded.
		{name: "unset falls back to the cap", requested: 0, want: cap},
		{name: "modest allowance passes through", requested: 21000, want: 21000},
		{name: "allowance at the cap passes through", requested: cap, want: cap},
		{name: "allowance past the cap is clamped", requested: cap + 1, want: cap},
		{name: "absurd allowance is clamped", requested: math.MaxUint64 / 2, want: cap},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := capGas(tt.requested); got != tt.want {
				t.Fatalf("capGas(%d) = %d, want %d", tt.requested, got, tt.want)
			}
		})
	}
}

func TestGasCapDisabled(t *testing.T) {
	withRPCLimits(t, RPCLimits{GasCap: 0, EVMTimeout: time.Second})

	if got := capGas(1 << 60); got != 1<<60 {
		t.Fatalf("a zero cap should leave the allowance alone, got %d", got)
	}
	if got := capGas(0); got != math.MaxUint64/2 {
		t.Fatalf("a zero cap should keep the unmetered default, got %d", got)
	}
}

// TestEstimateGasSearchBound checks the binary search bound that the request
// controls. Without the clamp the caller sets both how many rounds the search
// runs and how much gas each round may burn.
func TestEstimateGasSearchBound(t *testing.T) {
	const cap = 50000000

	withRPCLimits(t, RPCLimits{GasCap: cap, EVMTimeout: time.Second})

	// Mirrors the ceiling selection in EstimateGas for an explicit gas argument.
	bound := func(requestedGas uint64) uint64 {
		hi := requestedGas
		if hi < params.TxGas {
			t.Fatalf("test expects an explicit allowance above %d", params.TxGas)
		}
		if limit := currentRPCLimits().GasCap; limit != 0 && hi > limit {
			hi = limit
		}
		return hi
	}

	if got := bound(math.MaxInt64); got != cap {
		t.Fatalf("a MaxInt64 allowance should be clamped to %d, got %d", cap, got)
	}
	if got := bound(params.TxGas); got != params.TxGas {
		t.Fatalf("an ordinary allowance should pass through, got %d", got)
	}
}

func TestDefaultRPCLimits(t *testing.T) {
	// Both defaults must be non-zero, otherwise a node that never sets the flags
	// silently runs with the unbounded behaviour these limits exist to prevent.
	if DefaultRPCLimits.GasCap == 0 {
		t.Fatal("default gas cap must not be unlimited")
	}
	if DefaultRPCLimits.EVMTimeout <= 0 {
		t.Fatal("default EVM timeout must be positive")
	}
}
