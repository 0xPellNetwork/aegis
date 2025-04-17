# tx group

Group transaction subcommands

```
pellcored tx group [flags]
```

### Options

```
  -h, --help   help for group
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID
      --home string         directory for config and data 
      --log_format string   The logging format (json|plain) 
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) 
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [pellcored tx](pellcored_tx.md)	 - Transactions subcommands
* [pellcored tx group create-group](pellcored_tx_group_create-group.md)	 - Create a group which is an aggregation of member accounts with associated weights and an administrator account.
* [pellcored tx group create-group-policy](pellcored_tx_group_create-group-policy.md)	 - Create a group policy which is an account associated with a group and a decision policy. Note, the '--from' flag is ignored as it is implied from [admin].
* [pellcored tx group create-group-with-policy](pellcored_tx_group_create-group-with-policy.md)	 - Create a group with policy which is an aggregation of member accounts with associated weights, an administrator account and decision policy.
* [pellcored tx group draft-proposal](pellcored_tx_group_draft-proposal.md)	 - Generate a draft proposal json file. The generated proposal json contains only one message (skeleton).
* [pellcored tx group exec](pellcored_tx_group_exec.md)	 - Execute a proposal
* [pellcored tx group leave-group](pellcored_tx_group_leave-group.md)	 - Remove member from the group
* [pellcored tx group submit-proposal](pellcored_tx_group_submit-proposal.md)	 - Submit a new proposal
* [pellcored tx group update-group-admin](pellcored_tx_group_update-group-admin.md)	 - Update a group's admin
* [pellcored tx group update-group-members](pellcored_tx_group_update-group-members.md)	 - Update a group's members. Set a member's weight to "0" to delete it.
* [pellcored tx group update-group-metadata](pellcored_tx_group_update-group-metadata.md)	 - Update a group's metadata
* [pellcored tx group update-group-policy-admin](pellcored_tx_group_update-group-policy-admin.md)	 - Update a group policy admin
* [pellcored tx group update-group-policy-decision-policy](pellcored_tx_group_update-group-policy-decision-policy.md)	 - Update a group policy's decision policy
* [pellcored tx group update-group-policy-metadata](pellcored_tx_group_update-group-policy-metadata.md)	 - Update a group policy metadata
* [pellcored tx group vote](pellcored_tx_group_vote.md)	 - Vote on a proposal
* [pellcored tx group withdraw-proposal](pellcored_tx_group_withdraw-proposal.md)	 - Withdraw a submitted proposal

