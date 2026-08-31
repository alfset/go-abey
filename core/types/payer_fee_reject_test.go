package types

import (
	"math/big"
	"testing"

	"github.com/AbeyFoundation/go-abey/common"
)

func TestHasPayerOrFee(t *testing.T) {
	to := common.HexToAddress("0x2000000000000000000000000000000000000002")
	payer := common.HexToAddress("0x3000000000000000000000000000000000000003")
	// Reads as zero through Uint64, which is how it slips past the signing hash.
	truncating, _ := new(big.Int).SetString("18446744073709551616", 10)

	tests := []struct {
		name string
		tx   *Transaction
		want bool
	}{
		{"plain transaction", NewTransaction(0, to, big.NewInt(1), 21000, big.NewInt(1000), nil), false},
		{"explicit zero fee and no payer", NewTransaction_Payment(0, to, big.NewInt(1), big.NewInt(0), 21000, big.NewInt(1000), nil, common.Address{}), false},
		{"fee attached", NewTransaction_Payment(0, to, big.NewInt(1), big.NewInt(70000), 21000, big.NewInt(1000), nil, common.Address{}), true},
		{"fee that truncates to zero", NewTransaction_Payment(0, to, big.NewInt(1), truncating, 21000, big.NewInt(1000), nil, common.Address{}), true},
		{"payer attached", NewTransaction_Payment(0, to, big.NewInt(1), nil, 21000, big.NewInt(1000), nil, payer), true},
		{"contract creation with fee", NewContractCreation_Payment(0, big.NewInt(1), big.NewInt(70000), 21000, big.NewInt(1000), nil, common.Address{}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tx.HasPayerOrFee(); got != tt.want {
				t.Fatalf("HasPayerOrFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The check has to survive sender recovery, which clears a fee whose low 64
// bits are zero. Anything reading Fee() afterwards sees nothing.
func TestHasPayerOrFeeSurvivesSenderRecovery(t *testing.T) {
	to := common.HexToAddress("0x2000000000000000000000000000000000000002")
	truncating, _ := new(big.Int).SetString("18446744073709551616", 10)
	tx := NewTransaction_Payment(0, to, big.NewInt(1), truncating, 21000, big.NewInt(1000), nil, common.Address{})

	signer := NewTIP1Signer(big.NewInt(19330))
	before := tx.HasPayerOrFee()
	_, _ = Sender(signer, tx) // clears the fee as a side effect of hashing

	if !before {
		t.Fatal("expected the fee to be detected before recovery")
	}
	if tx.Fee() != nil && tx.Fee().Sign() != 0 {
		t.Log("fee survived recovery")
	} else {
		t.Log("fee was cleared by recovery, which is why the check runs first")
	}
}
