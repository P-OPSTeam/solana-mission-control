package exporter

import (
	"encoding/json"
	"testing"

	"github.com/Chainflow/solana-mission-control/types"
)

func TestEpochVoteCredits(t *testing.T) {
	body := `{"result":{"current":[
		{"nodePubkey":"a","epochCredits":[[499,900,800],[500,1000,900]]},
		{"nodePubkey":"b","epochCredits":[[500,18446744073709551615,18446744073709551615]]}
	],"delinquent":[]}}`
	var r types.GetVoteAccountsResponse
	if err := json.Unmarshal([]byte(body), &r); err != nil {
		t.Fatal(err)
	}
	c, p := epochVoteCredits(r.Result.Current[0].EpochCredits, 500)
	if c != "1000" || p != "900" {
		t.Fatalf("got %s %s", c, p)
	}
	c, p = epochVoteCredits(r.Result.Current[1].EpochCredits, 500)
	if c != "" || p != "" {
		t.Fatalf("sentinel not skipped: %q %q", c, p)
	}
	c, p = epochVoteCredits(r.Result.Current[0].EpochCredits, 7)
	if c != "0" || p != "0" {
		t.Fatalf("missing epoch: %s %s", c, p)
	}
}
