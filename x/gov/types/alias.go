package types

import (
	sdkGov "github.com/cosmos/cosmos-sdk/x/gov"
	sdkGovTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// nolint
	EventTypeSubmitProposal   = sdkGovTypes.EventTypeSubmitProposal
	EventTypeProposalDeposit  = sdkGovTypes.EventTypeProposalDeposit
	EventTypeProposalVote     = sdkGovTypes.EventTypeProposalVote
	EventTypeInactiveProposal = sdkGovTypes.EventTypeInactiveProposal
	EventTypeActiveProposal   = sdkGovTypes.EventTypeActiveProposal

	AttributeKeyProposalResult     = sdkGovTypes.AttributeKeyProposalResult
	AttributeKeyOption             = sdkGovTypes.AttributeKeyOption
	AttributeKeyProposalID         = sdkGovTypes.AttributeKeyProposalID
	AttributeKeyVotingPeriodStart  = sdkGovTypes.AttributeKeyVotingPeriodStart
	AttributeValueCategory         = sdkGovTypes.AttributeValueCategory
	AttributeValueProposalDropped  = sdkGovTypes.AttributeValueProposalDropped
	AttributeValueProposalPassed   = sdkGovTypes.AttributeValueProposalPassed
	AttributeValueProposalRejected = sdkGovTypes.AttributeValueProposalRejected
	AttributeValueProposalFailed   = sdkGovTypes.AttributeValueProposalFailed

	ModuleName = sdkGovTypes.ModuleName

	StatusNil           = sdkGovTypes.StatusNil
	StatusDepositPeriod = sdkGovTypes.StatusDepositPeriod
	StatusVotingPeriod  = sdkGovTypes.StatusVotingPeriod
	StatusPassed        = sdkGovTypes.StatusPassed
	StatusRejected      = sdkGovTypes.StatusRejected
	StatusFailed        = sdkGovTypes.StatusFailed

	OptionEmpty      = sdkGovTypes.OptionEmpty
	OptionYes        = sdkGovTypes.OptionYes
	OptionAbstain    = sdkGovTypes.OptionAbstain
	OptionNo         = sdkGovTypes.OptionNo
	OptionNoWithVeto = sdkGovTypes.OptionNoWithVeto
	MaxTitleLength   = sdkGovTypes.MaxTitleLength

	StoreKey              = sdkGovTypes.StoreKey
	RouterKey             = sdkGovTypes.RouterKey
	QuerierRoute          = sdkGovTypes.QuerierRoute
	DefaultParamspace     = sdkGovTypes.ModuleName
	TypeMsgDeposit        = sdkGovTypes.TypeMsgDeposit
	TypeMsgVote           = sdkGovTypes.TypeMsgVote
	TypeMsgSubmitProposal = sdkGovTypes.TypeMsgSubmitProposal
	ProposalTypeText      = sdkGovTypes.ProposalTypeText

	QueryParams    = sdkGovTypes.QueryParams
	QueryProposals = sdkGovTypes.QueryProposals
	QueryProposal  = sdkGovTypes.QueryProposal
	QueryDeposits  = sdkGovTypes.QueryDeposits
	QueryDeposit   = sdkGovTypes.QueryDeposit
	QueryVotes     = sdkGovTypes.QueryVotes
	QueryVote      = sdkGovTypes.QueryVote
	QueryTally     = sdkGovTypes.QueryTally

	ParamDeposit  = sdkGovTypes.ParamDeposit
	ParamVoting   = sdkGovTypes.ParamVoting
	ParamTallying = sdkGovTypes.ParamTallying

	MaxDescriptionLength = sdkGovTypes.MaxDescriptionLength
)

var (
	// nolint
	ErrNoProposalHandlerExists = sdkGovTypes.ErrNoProposalHandlerExists
	ErrInvalidProposalContent  = sdkGovTypes.ErrInvalidProposalContent
	ErrInvalidGenesis          = sdkGovTypes.ErrInvalidGenesis
	ErrInvalidProposalType     = sdkGovTypes.ErrInvalidProposalType
	ErrInvalidVote             = sdkGovTypes.ErrInvalidVote

	ProposalKey         = sdkGovTypes.ProposalKey
	ValidProposalStatus = sdkGovTypes.ValidProposalStatus

	ProposalIDKey = sdkGovTypes.ProposalIDKey

	DepositsKey = sdkGovTypes.DepositsKey
	VotesKey    = sdkGovTypes.VotesKey

	ProposalsKeyPrefix          = sdkGovTypes.ProposalsKeyPrefix
	DepositsKeyPrefix           = sdkGovTypes.DepositsKeyPrefix
	VotesKeyPrefix              = sdkGovTypes.VotesKeyPrefix
	ActiveProposalQueuePrefix   = sdkGovTypes.ActiveProposalQueuePrefix
	InactiveProposalQueuePrefix = sdkGovTypes.InactiveProposalQueuePrefix
	ValidVoteOption             = sdkGovTypes.ValidVoteOption

	ParamKeyTable = sdkGovTypes.ParamKeyTable

	ParamStoreKeyDepositParams = sdkGovTypes.ParamStoreKeyDepositParams
	ParamStoreKeyVotingParams  = sdkGovTypes.ParamStoreKeyVotingParams
	ParamStoreKeyTallyParams   = sdkGovTypes.ParamStoreKeyTallyParams

	NewAppModule = sdkGov.NewAppModule

	NewTallyResultFromMap     = sdkGovTypes.NewTallyResultFromMap
	EmptyTallyResult          = sdkGovTypes.EmptyTallyResult
	RegisterProposalType      = sdkGovTypes.RegisterProposalType
	RegisterProposalTypeCodec = sdkGovTypes.RegisterProposalTypeCodec
	RegisterCodec             = sdkGovTypes.RegisterCodec

	ActiveProposalByTimeKey   = sdkGovTypes.ActiveProposalByTimeKey
	ActiveProposalQueueKey    = sdkGovTypes.ActiveProposalQueueKey
	InactiveProposalByTimeKey = sdkGovTypes.InactiveProposalByTimeKey
	InactiveProposalQueueKey  = sdkGovTypes.InactiveProposalQueueKey

	NewMsgSubmitProposal    = sdkGovTypes.NewMsgSubmitProposal
	NewMsgDeposit           = sdkGovTypes.NewMsgDeposit
	NewMsgVote              = sdkGovTypes.NewMsgVote
	NewDepositParams        = sdkGovTypes.NewDepositParams
	NewTallyParams          = sdkGovTypes.NewTallyParams
	NewVotingParams         = sdkGovTypes.NewVotingParams
	NewParams               = sdkGovTypes.NewParams
	NewTextProposal         = sdkGovTypes.NewTextProposal
	ContentFromProposalType = sdkGovTypes.ContentFromProposalType
	IsValidProposalType     = sdkGovTypes.IsValidProposalType
	ProposalHandler         = sdkGovTypes.ProposalHandler
	ModuleCdc               = sdkGovTypes.ModuleCdc

	ValidateAbstract            = sdkGovTypes.ValidateAbstract
	ProposalStatusFromString    = sdkGovTypes.ProposalStatusFromString
	VoteOptionFromString        = sdkGovTypes.VoteOptionFromString
	NewQueryVoteParams          = sdkGovTypes.NewQueryVoteParams
	NewQueryProposalParams      = sdkGovTypes.NewQueryProposalParams
	NewQueryProposalVotesParams = sdkGovTypes.NewQueryProposalVotesParams
	NewQueryDepositParams       = sdkGovTypes.NewQueryDepositParams

	NewQueryProposalsParams = sdkGovTypes.NewQueryProposalsParams
	VoteKey                 = sdkGovTypes.VoteKey
	DepositKey              = sdkGovTypes.DepositKey
	ContentToAny            = sdkGovTypes.ContentToAny

	GetProposalIDFromBytes = sdkGovTypes.GetProposalIDFromBytes
	GetProposalIDBytes     = sdkGovTypes.GetProposalIDBytes
)

type (
	// nolint
	ProposalStatus       = sdkGovTypes.ProposalStatus
	VoteOption           = sdkGovTypes.VoteOption
	Vote                 = sdkGovTypes.Vote
	Votes                = sdkGovTypes.Votes
	Deposit              = sdkGovTypes.Deposit
	Deposits             = sdkGovTypes.Deposits
	DepositParams        = sdkGovTypes.DepositParams
	VotingParams         = sdkGovTypes.VotingParams
	TallyParams          = sdkGovTypes.TallyParams
	Params               = sdkGovTypes.Params
	Proposal             = sdkGovTypes.Proposal
	Proposals            = sdkGovTypes.Proposals
	Content              = sdkGovTypes.Content
	TextProposal         = sdkGovTypes.TextProposal
	TallyResult          = sdkGovTypes.TallyResult
	Handler              = sdkGovTypes.Handler
	MsgSubmitProposal    = sdkGovTypes.MsgSubmitProposal
	MsgDeposit           = sdkGovTypes.MsgDeposit
	MsgVote              = sdkGovTypes.MsgVote
	QueryProposalParams  = sdkGovTypes.QueryProposalParams
	QueryDepositParams   = sdkGovTypes.QueryDepositParams
	QueryVoteParams      = sdkGovTypes.QueryVoteParams
	QueryProposalsParams = sdkGovTypes.QueryProposalsParams
	BankKeeper           = sdkGovTypes.BankKeeper
)
