package monitor

import (
	"encoding/json"
	"fmt"
	"log"
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

// return validator epochCredit and networkEpochCredit
func GetEpochCredits(cfg *config.Config) (float64, float64, error) {
	log.Println("Getting Epoch Credit...")
	var countNonZeroEpochCreditValidator int
	var valEpochCredit, netEpochCredit, avgnetEpochCredit float64

	if solanaBinaryPath == "" {
		solanaBinaryPath = "solana"
	}

	log.Printf("Solana binary path : %s", solanaBinaryPath)

	cmd := exec.Command(solanaBinaryPath, "validators", "--output", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error while running solana validators cli command %v", err)
		return valEpochCredit, avgnetEpochCredit, err
	}

	var result types.SkipRate
	err = json.Unmarshal(out, &result)
	if err != nil {
		log.Printf("Error: %v", err)
		return valEpochCredit, avgnetEpochCredit, err
	}

	for _, val := range result.Validators {
		if val.IdentityPubkey == cfg.ValDetails.PubKey {
			valEpochCredit = float64(val.EpochCredits)
		}
		netEpochCredit = netEpochCredit + float64(val.EpochCredits)
		if (val.EpochCredits != 0) {
			countNonZeroEpochCreditValidator ++
		}
	}

	avgnetEpochCredit = netEpochCredit / float64(countNonZeroEpochCreditValidator)

	log.Printf("VAL epochCredit : %f, AVG Network epochCredit : %f", valEpochCredit, avgnetEpochCredit)

	return valEpochCredit, avgnetEpochCredit, nil
}
