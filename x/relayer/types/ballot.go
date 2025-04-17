package types

import (
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
)

func (m Ballot) AddVote(address string, vote VoteType) (Ballot, error) {
	if m.HasVoted(address) {
		return m, cosmoserrors.Wrap(ErrUnableToAddVote, fmt.Sprintf(" Voter : %s | Ballot :%s | Already Voted", address, m.String()))
	}
	// `index` is the index of the `address` in the `VoterList`
	// `index` is used to set the vote in the `Votes` array
	index := m.GetVoterIndex(address)
	if index == -1 {
		return m, cosmoserrors.Wrap(ErrUnableToAddVote, fmt.Sprintf("Voter %s not in voter list", address))
	}
	m.Votes[index] = vote
	return m, nil
}

func (m Ballot) HasVoted(address string) bool {
	index := m.GetVoterIndex(address)
	if index == -1 {
		return false
	}
	return m.Votes[index] != VoteType_NOT_YET_VOTED
}

// GetVoterIndex returns the index of the `address` in the `VoterList`
func (m Ballot) GetVoterIndex(address string) int {
	index := -1
	for i, addr := range m.VoterList {
		if addr == address {
			return i
		}
	}
	return index
}

// IsFinalizingVote checks sets the ballot to a final status if enough votes have been added
// If it has already been finalized it returns false
// It enough votes have not been added it returns false
func (m Ballot) IsFinalizingVote() (Ballot, bool) {
	if m.BallotStatus != BallotStatus_BALLOT_IN_PROGRESS {
		return m, false
	}
	success, failure := sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()
	total := sdkmath.LegacyNewDec(int64(len(m.VoterList)))
	if total.IsZero() {
		return m, false
	}
	for _, vote := range m.Votes {
		if vote == VoteType_SUCCESS_OBSERVATION {
			success = success.Add(sdkmath.LegacyOneDec())
		}
		if vote == VoteType_FAILURE_OBSERVATION {
			failure = failure.Add(sdkmath.LegacyOneDec())
		}

	}
	if failure.IsPositive() {
		if failure.Quo(total).GTE(m.BallotThreshold) {
			m.BallotStatus = BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION
			return m, true
		}
	}

	if success.IsPositive() {
		if success.Quo(total).GTE(m.BallotThreshold) {
			m.BallotStatus = BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION
			return m, true
		}
	}
	return m, false
}

func CreateVotes(len int) []VoteType {
	voterList := make([]VoteType, len)
	for i := range voterList {
		voterList[i] = VoteType_NOT_YET_VOTED
	}
	return voterList
}

// BuildRewardsDistribution builds the rewards distribution map for the ballot
// It returns the total rewards units which account for the observer block rewards
func (m Ballot) BuildRewardsDistribution(rewardsMap map[string]int64) int64 {
	totalRewardUnits := int64(0)
	switch m.BallotStatus {
	case BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION:
		for _, address := range m.VoterList {
			vote := m.Votes[m.GetVoterIndex(address)]
			if vote == VoteType_SUCCESS_OBSERVATION {
				rewardsMap[address] = rewardsMap[address] + 1
				totalRewardUnits++
				continue
			}
			rewardsMap[address] = rewardsMap[address] - 1
		}
	case BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION:
		for _, address := range m.VoterList {
			vote := m.Votes[m.GetVoterIndex(address)]
			if vote == VoteType_FAILURE_OBSERVATION {
				rewardsMap[address] = rewardsMap[address] + 1
				totalRewardUnits++
				continue
			}
			rewardsMap[address] = rewardsMap[address] - 1
		}
	}
	return totalRewardUnits
}
