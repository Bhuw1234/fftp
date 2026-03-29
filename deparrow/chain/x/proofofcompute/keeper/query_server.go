// Package keeper implements query handling for the proofofcompute module
package keeper

import (
	"encoding/json"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// QueryParamsResponse is the response for params query
type QueryParamsResponse struct {
	Params types.Params `json:"params"`
}

// QueryJobResponse is the response for job query
type QueryJobResponse struct {
	Job types.Job `json:"job"`
}

// QueryJobsResponse is the response for jobs list query
type QueryJobsResponse struct {
	Jobs []types.Job `json:"jobs"`
}

// QueryProofResponse is the response for proof query
type QueryProofResponse struct {
	Proof types.ComputeProof `json:"proof"`
}

// QueryDifficultyResponse is the response for difficulty query
type QueryDifficultyResponse struct {
	CurrentDifficulty  uint64 `json:"current_difficulty"`
	JobsInPeriod       uint64 `json:"jobs_in_period"`
	BlocksToAdjustment int64  `json:"blocks_to_adjustment"`
}

// QueryRewardStatsResponse is the response for reward stats query
type QueryRewardStatsResponse struct {
	TotalDistributed string `json:"total_distributed"`
	TotalJobs        uint64 `json:"total_jobs"`
	MaxSupply        string `json:"max_supply"`
}

// QueryPendingRewardResponse is the response for pending reward query
type QueryPendingRewardResponse struct {
	Amount types.Coin `json:"amount"`
}

// HandleQuery handles query requests
func (k Keeper) HandleQuery(path string, data []byte) ([]byte, error) {
	switch path {
	case "/proofofcompute/params":
		return k.queryParams()

	case "/proofofcompute/job":
		var req struct {
			JobID string `json:"job_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryJob(req.JobID)

	case "/proofofcompute/jobs":
		return k.queryJobs()

	case "/proofofcompute/jobs_by_submitter":
		var req struct {
			Submitter string `json:"submitter"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryJobsBySubmitter(req.Submitter)

	case "/proofofcompute/jobs_by_compute_node":
		var req struct {
			NodeAddress string `json:"node_address"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryJobsByComputeNode(req.NodeAddress)

	case "/proofofcompute/proof":
		var req struct {
			JobID string `json:"job_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryProof(req.JobID)

	case "/proofofcompute/difficulty":
		return k.queryDifficulty()

	case "/proofofcompute/reward_stats":
		return k.queryRewardStats()

	case "/proofofcompute/pending_reward":
		var req struct {
			NodeAddress string `json:"node_address"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryPendingReward(req.NodeAddress)

	case "/proofofcompute/stats":
		return k.queryStats()

	default:
		return nil, nil // Not handled by this module
	}
}

func (k Keeper) queryParams() ([]byte, error) {
	params := k.GetParams()
	resp := QueryParamsResponse{Params: params}
	return json.Marshal(resp)
}

func (k Keeper) queryJob(jobID string) ([]byte, error) {
	job, found := k.GetJob(jobID)
	if !found {
		return nil, types.ErrJobNotFound
	}
	resp := QueryJobResponse{Job: job}
	return json.Marshal(resp)
}

func (k Keeper) queryJobs() ([]byte, error) {
	jobs := k.GetAllJobs()
	resp := QueryJobsResponse{Jobs: jobs}
	return json.Marshal(resp)
}

func (k Keeper) queryJobsBySubmitter(submitter string) ([]byte, error) {
	jobs := k.GetJobsBySubmitter(submitter)
	resp := QueryJobsResponse{Jobs: jobs}
	return json.Marshal(resp)
}

func (k Keeper) queryJobsByComputeNode(nodeAddress string) ([]byte, error) {
	jobs := k.GetJobsByComputeNode(nodeAddress)
	resp := QueryJobsResponse{Jobs: jobs}
	return json.Marshal(resp)
}

func (k Keeper) queryProof(jobID string) ([]byte, error) {
	proof, found := k.GetProof(jobID)
	if !found {
		return nil, types.ErrJobNotFound
	}
	resp := QueryProofResponse{Proof: proof}
	return json.Marshal(resp)
}

func (k Keeper) queryDifficulty() ([]byte, error) {
	resp := QueryDifficultyResponse{
		CurrentDifficulty:  k.GetCurrentDifficulty(),
		JobsInPeriod:       0, // TODO: track this
		BlocksToAdjustment: 0, // TODO: track this
	}
	return json.Marshal(resp)
}

func (k Keeper) queryRewardStats() ([]byte, error) {
	stats := k.GetStats()
	jobs := k.GetAllJobs()
	var completed uint64
	for _, job := range jobs {
		if job.Status == types.JobStatusCompleted {
			completed++
		}
	}

	resp := QueryRewardStatsResponse{
		TotalDistributed: stats["total_supply"].(string),
		TotalJobs:        completed,
		MaxSupply:        k.GetParams().MaxSupply,
	}
	return json.Marshal(resp)
}

func (k Keeper) queryPendingReward(nodeAddress string) ([]byte, error) {
	reward, found := k.GetPendingReward(nodeAddress)
	if !found {
		return json.Marshal(QueryPendingRewardResponse{
			Amount: types.Coin{Denom: "dpc", Amount: "0"},
		})
	}
	resp := QueryPendingRewardResponse{
		Amount: types.Coin{Denom: "dpc", Amount: reward.Amount},
	}
	return json.Marshal(resp)
}

func (k Keeper) queryStats() ([]byte, error) {
	stats := k.GetStats()
	return json.Marshal(stats)
}