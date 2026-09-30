package monitor

import (
	"math"
	"testing"

	"github.com/Chainflow/solana-mission-control/types"
)

func TestCalcEpochCreditStats(t *testing.T) {
	v := []types.SkipRateValidator{
		{IdentityPubkey: "me", EpochCredits: 60, ActivatedStake: 100e9},
		{IdentityPubkey: "big", EpochCredits: 600, ActivatedStake: 1000e9},
		{IdentityPubkey: "none", EpochCredits: math.MaxUint64, ActivatedStake: 50e9},
		{IdentityPubkey: "zero", EpochCredits: 0, ActivatedStake: 10e9},
	}
	s := calcEpochCreditStats(v, "me")
	if s.Validator != 60 || s.Network != 330 {
		t.Fatalf("got %+v", s)
	}
	if s.ValidatorPerStake != 0.6 || s.NetworkPerStake != 0.6 {
		t.Fatalf("got %+v", s)
	}
}
