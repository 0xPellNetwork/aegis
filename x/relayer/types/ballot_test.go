package types

import (
	"testing"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestBallot_AddVote(t *testing.T) {
	type votes struct {
		address string
		vote    VoteType
	}

	tt := []struct {
		name        string
		threshold   math.LegacyDec
		voterList   []string
		votes       []votes
		finalVotes  []VoteType
		finalStatus BallotStatus
		isFinalized bool
		wantErr     bool
	}{
		{
			name:      "All success",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer2", VoteType_SUCCESS_OBSERVATION},
				{"Observer3", VoteType_SUCCESS_OBSERVATION},
				{"Observer4", VoteType_SUCCESS_OBSERVATION},
			},
			finalVotes:  []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION},
			finalStatus: BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			isFinalized: true,
		},
		{
			name:      "Multiple votes by a observer , Ballot success",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer1", VoteType_FAILURE_OBSERVATION},
				{"Observer2", VoteType_SUCCESS_OBSERVATION},
				{"Observer3", VoteType_SUCCESS_OBSERVATION},
				{"Observer4", VoteType_SUCCESS_OBSERVATION},
			},
			finalVotes:  []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION},
			finalStatus: BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			isFinalized: true,
		},
		{
			name:      "Multiple votes by a observer , Ballot in progress",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer1", VoteType_FAILURE_OBSERVATION},
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
			},
			finalVotes:  []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_NOT_YET_VOTED, VoteType_NOT_YET_VOTED, VoteType_NOT_YET_VOTED},
			finalStatus: BallotStatus_BALLOT_IN_PROGRESS,
			isFinalized: false,
		},
		{
			name:      "Ballot finalized at threshold",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4", "Observer5", "Observer6", "Observer7", "Observer8", "Observer9", "Observer10", "Observer11", "Observer12"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer2", VoteType_SUCCESS_OBSERVATION},
				{"Observer3", VoteType_SUCCESS_OBSERVATION},
				{"Observer4", VoteType_SUCCESS_OBSERVATION},
				{"Observer5", VoteType_SUCCESS_OBSERVATION},
				{"Observer6", VoteType_SUCCESS_OBSERVATION},
				{"Observer7", VoteType_SUCCESS_OBSERVATION},
				{"Observer8", VoteType_SUCCESS_OBSERVATION},
				{"Observer9", VoteType_NOT_YET_VOTED},
				{"Observer10", VoteType_NOT_YET_VOTED},
				{"Observer11", VoteType_NOT_YET_VOTED},
				{"Observer12", VoteType_NOT_YET_VOTED},
			},
			finalVotes: []VoteType{VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
			},
			finalStatus: BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			isFinalized: true,
		},
		{
			name:      "Ballot finalized at threshold but more votes added after",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4", "Observer5", "Observer6", "Observer7", "Observer8", "Observer9", "Observer10", "Observer11", "Observer12"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer2", VoteType_SUCCESS_OBSERVATION},
				{"Observer3", VoteType_SUCCESS_OBSERVATION},
				{"Observer4", VoteType_SUCCESS_OBSERVATION},
				{"Observer5", VoteType_SUCCESS_OBSERVATION},
				{"Observer6", VoteType_SUCCESS_OBSERVATION},
				{"Observer7", VoteType_SUCCESS_OBSERVATION},
				{"Observer8", VoteType_SUCCESS_OBSERVATION},
				{"Observer9", VoteType_SUCCESS_OBSERVATION},
				{"Observer10", VoteType_SUCCESS_OBSERVATION},
				{"Observer11", VoteType_SUCCESS_OBSERVATION},
				{"Observer12", VoteType_SUCCESS_OBSERVATION},
			},
			finalVotes: []VoteType{VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
			},
			finalStatus: BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			isFinalized: true,
		},
		{
			name:      "Two observers",
			threshold: sdkmath.LegacyMustNewDecFromStr("1.00"),
			voterList: []string{"Observer1", "Observer2"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer2", VoteType_SUCCESS_OBSERVATION},
			},
			finalVotes:  []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION},
			finalStatus: BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			isFinalized: true,
		},
		{
			name:      "Low threshold 1 always fails as Failure is checked first",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.01"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer2", VoteType_SUCCESS_OBSERVATION},
				{"Observer3", VoteType_SUCCESS_OBSERVATION},
				{"Observer4", VoteType_FAILURE_OBSERVATION},
			},
			finalVotes:  []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_FAILURE_OBSERVATION},
			finalStatus: BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION,
			isFinalized: true,
		},
		{
			name:      "Low threshold 2 always fails as Failure is checked first",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.01"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer2", VoteType_FAILURE_OBSERVATION},
				{"Observer3", VoteType_SUCCESS_OBSERVATION},
				{"Observer4", VoteType_SUCCESS_OBSERVATION},
			},
			finalVotes:  []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_FAILURE_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION},
			finalStatus: BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION,
			isFinalized: true,
		},
		{
			name:      "100 percent threshold cannot finalze with less than 100 percent votes",
			threshold: sdkmath.LegacyMustNewDecFromStr("1"),
			voterList: []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes: []votes{
				{"Observer1", VoteType_SUCCESS_OBSERVATION},
				{"Observer2", VoteType_FAILURE_OBSERVATION},
				{"Observer3", VoteType_SUCCESS_OBSERVATION},
				{"Observer4", VoteType_SUCCESS_OBSERVATION},
			},
			finalVotes:  []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_FAILURE_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION},
			finalStatus: BallotStatus_BALLOT_IN_PROGRESS,
			isFinalized: false,
		},
		{
			name:      "Voter not in voter list",
			threshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			voterList: []string{},
			votes: []votes{
				{"Observer5", VoteType_SUCCESS_OBSERVATION},
			},
			wantErr:     true,
			finalVotes:  []VoteType{},
			finalStatus: BallotStatus_BALLOT_IN_PROGRESS,
			isFinalized: false,
		},
	}
	for _, test := range tt {
		test := test
		t.Run(test.name, func(t *testing.T) {
			ballot := Ballot{
				Index:            "index",
				BallotIdentifier: "identifier",
				VoterList:        test.voterList,
				Votes:            CreateVotes(len(test.voterList)),
				ObservationType:  ObservationType_IN_BOUND_TX,
				BallotThreshold:  test.threshold,
				BallotStatus:     BallotStatus_BALLOT_IN_PROGRESS,
			}
			for _, vote := range test.votes {
				b, err := ballot.AddVote(vote.address, vote.vote)
				if test.wantErr {
					require.Error(t, err)
				}
				ballot = b
			}

			finalBallot, isFinalized := ballot.IsFinalizingVote()
			require.Equal(t, test.finalStatus, finalBallot.BallotStatus)
			require.Equal(t, test.finalVotes, finalBallot.Votes)
			require.Equal(t, test.isFinalized, isFinalized)
		})
	}
}

func TestBallot_IsFinalizingVote(t *testing.T) {
	tt := []struct {
		name            string
		BallotThreshold math.LegacyDec
		Votes           []VoteType
		finalizingVote  int
		finalStatus     BallotStatus
	}{
		{
			name:            "finalized to success",
			BallotThreshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			Votes: []VoteType{
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
			},
			finalizingVote: 7,
			finalStatus:    BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
		},
		{
			name:            "finalized to failure",
			BallotThreshold: sdkmath.LegacyMustNewDecFromStr("0.66"),
			Votes: []VoteType{
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
			},
			finalizingVote: 7,
			finalStatus:    BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION,
		},
		{
			name:            "low threshold finalized early to success",
			BallotThreshold: sdkmath.LegacyMustNewDecFromStr("0.01"),
			Votes: []VoteType{
				VoteType_SUCCESS_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
				VoteType_NOT_YET_VOTED,
			},
			finalizingVote: 0,
			finalStatus:    BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
		},
		{
			name:            "100 percent threshold cannot finalize with less than 100 percent votes",
			BallotThreshold: sdkmath.LegacyMustNewDecFromStr("1"),
			Votes: []VoteType{
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_SUCCESS_OBSERVATION,
			},
			finalizingVote: 0,
			finalStatus:    BallotStatus_BALLOT_IN_PROGRESS,
		},
		{
			name:            "100 percent threshold can finalize with 100 percent votes",
			BallotThreshold: sdkmath.LegacyMustNewDecFromStr("1"),
			Votes: []VoteType{
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
				VoteType_FAILURE_OBSERVATION,
			},
			finalizingVote: 11,
			finalStatus:    BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION,
		},
	}
	for _, test := range tt {
		test := test
		t.Run(test.name, func(t *testing.T) {

			ballot := Ballot{
				BallotStatus:    BallotStatus_BALLOT_IN_PROGRESS,
				BallotThreshold: test.BallotThreshold,
				VoterList:       make([]string, len(test.Votes)),
			}
			isFinalizingVote := false
			for index, vote := range test.Votes {
				ballot.Votes = append(ballot.Votes, vote)
				ballot, isFinalizingVote = ballot.IsFinalizingVote()
				if isFinalizingVote {
					require.Equal(t, test.finalizingVote, index)
				}
			}
			require.Equal(t, test.finalStatus, ballot.BallotStatus)
		})
	}
}

func Test_BuildRewardsDistribution(t *testing.T) {
	tt := []struct {
		name         string
		voterList    []string
		votes        []VoteType
		ballotStatus BallotStatus
		expectedMap  map[string]int64
	}{
		{
			name:         "BallotFinalized_SuccessObservation",
			voterList:    []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes:        []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_FAILURE_OBSERVATION},
			ballotStatus: BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			expectedMap: map[string]int64{
				"Observer1": 1,
				"Observer2": 1,
				"Observer3": 1,
				"Observer4": -1,
			},
		},
		{
			name:         "BallotFinalized_FailureObservation",
			voterList:    []string{"Observer1", "Observer2", "Observer3", "Observer4"},
			votes:        []VoteType{VoteType_SUCCESS_OBSERVATION, VoteType_SUCCESS_OBSERVATION, VoteType_FAILURE_OBSERVATION, VoteType_FAILURE_OBSERVATION},
			ballotStatus: BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION,
			expectedMap: map[string]int64{
				"Observer1": -1,
				"Observer2": -1,
				"Observer3": 1,
				"Observer4": 1,
			},
		},
	}
	for _, test := range tt {
		test := test
		t.Run(test.name, func(t *testing.T) {
			ballot := Ballot{
				Index:            "",
				BallotIdentifier: "",
				VoterList:        test.voterList,
				Votes:            test.votes,
				ObservationType:  0,
				BallotThreshold:  math.LegacyDec{},
				BallotStatus:     test.ballotStatus,
			}
			rewardsMap := map[string]int64{}
			ballot.BuildRewardsDistribution(rewardsMap)
			require.Equal(t, test.expectedMap, rewardsMap)
		})
	}

}
