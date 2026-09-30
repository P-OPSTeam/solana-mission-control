package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os/exec"

	"github.com/Chainflow/solana-mission-control/config"
	"github.com/Chainflow/solana-mission-control/types"
	"github.com/Chainflow/solana-mission-control/utils"
)

// GetVoteAccounts returns voting accounts information
func GetVoteAccounts(cfg *config.Config, node string) (types.GetVoteAccountsResponse, error) {
	log.Println("Getting Vote Account Information...")
	ops := types.HTTPOptions{
		Endpoint: cfg.Endpoints.RPCEndpoint,
		Method:   http.MethodPost,
		Body: types.Payload{Jsonrpc: "2.0", Method: "getVoteAccounts", ID: 1, Params: []interface{}{
			types.Commitment{
				Commitemnt: "recent",
			},
		}},
	}
	if node == utils.Network {
		ops.Endpoint = cfg.Endpoints.NetworkRPC
	} else if node == utils.Validator {
		ops.Endpoint = cfg.Endpoints.RPCEndpoint
	} else {
		ops.Endpoint = cfg.Endpoints.RPCEndpoint
	}

	var result types.GetVoteAccountsResponse

	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("GetVoteAccounts - Error while getting leader shedules: %v", err)
		return result, err
	}

	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		log.Printf("Error while unmarshelling leader shedules: %v", err)
		return result, err
	}

	if result.Error.Code != 0 {
		return result, fmt.Errorf("RPC error: %d %v", result.Error.Code, result.Error.Message)
	}

	return result, nil
}

// EpochCreditStats holds epoch credit values of the validator and the network.
// The PerStake values are credits per SOL of activated stake. After Alpenglow
// credits are stake weighted, so the plain network average is not comparable
// to a single validator.
type EpochCreditStats struct {
	Validator          float64
	Network            float64
	ValidatorPerStake  float64
	NetworkPerStake    float64
}

// calcEpochCreditStats computes epoch credit stats from the validators list.
// Validators reporting u64::MAX (credits not available) are ignored.
func calcEpochCreditStats(validators []types.SkipRateValidator, pubKey string) EpochCreditStats {
	var stats EpochCreditStats
	var count int
	var netCredits, netStake float64

	for _, val := range validators {
		if val.EpochCredits == math.MaxUint64 {
			continue
		}
		credits := float64(val.EpochCredits)
		stake := float64(val.ActivatedStake) / 1e9
		if val.IdentityPubkey == pubKey {
			stats.Validator = credits
			if stake > 0 {
				stats.ValidatorPerStake = credits / stake
			}
		}
		stats.Network += credits
		if val.EpochCredits != 0 {
			count++
			if stake > 0 {
				netCredits += credits
				netStake += stake
			}
		}
	}

	stats.Network = stats.Network / float64(count)
	if netStake > 0 {
		stats.NetworkPerStake = netCredits / netStake
	}
	return stats
}

// GetEpochCredits returns validator and network epoch credit stats
func GetEpochCredits(cfg *config.Config) (EpochCreditStats, error) {
	log.Println("Getting Epoch Credit...")

	if solanaBinaryPath == "" {
		solanaBinaryPath = "solana"
	}

	log.Printf("Solana binary path : %s", solanaBinaryPath)

	cmd := exec.Command(solanaBinaryPath, "validators", "--output", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error while running solana validators cli command %v", err)
		return EpochCreditStats{}, err
	}

	var result types.SkipRate
	err = json.Unmarshal(out, &result)
	if err != nil {
		log.Printf("Error: %v", err)
		return EpochCreditStats{}, err
	}

	stats := calcEpochCreditStats(result.Validators, cfg.ValDetails.PubKey)

	log.Printf("VAL epochCredit : %f, AVG Network epochCredit : %f, VAL per SOL : %f, NET per SOL : %f",
		stats.Validator, stats.Network, stats.ValidatorPerStake, stats.NetworkPerStake)

	return stats, nil
}
