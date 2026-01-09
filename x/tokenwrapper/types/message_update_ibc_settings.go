package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgFundModuleWallet{}

func NewMsgUpdateIbcSettings(signer string, nativeClientId string, counterpartyClientId string, nativePort string, counterpartyPort string, nativeChannel string, counterpartyChannel string, denom string, decimalDifference uint32) *MsgUpdateIbcSettings {
	return &MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       nativeClientId,
		CounterpartyClientId: counterpartyClientId,
		NativePort:           nativePort,
		CounterpartyPort:     counterpartyPort,
		NativeChannel:        nativeChannel,
		CounterpartyChannel:  counterpartyChannel,
		Denom:                denom,
		DecimalDifference:    decimalDifference,
	}
}

func (msg *MsgUpdateIbcSettings) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.ValidateClientId(msg.NativeClientId); err != nil {
		return err
	}

	if err := validators.ValidateClientId(msg.CounterpartyClientId); err != nil {
		return err
	}

	if err := validators.ValidatePort(msg.NativePort); err != nil {
		return err
	}

	if err := validators.ValidatePort(msg.CounterpartyPort); err != nil {
		return err
	}

	if err := validators.ValidateChannel(msg.NativeChannel); err != nil {
		return err
	}

	if err := validators.ValidateChannel(msg.CounterpartyChannel); err != nil {
		return err
	}

	if err := validators.ValidateDenom(msg.Denom); err != nil {
		return err
	}

	if err := validators.ValidateDecimalDifference(msg.DecimalDifference); err != nil {
		return err
	}

	return nil
}
