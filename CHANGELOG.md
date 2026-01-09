# Changelog
All notable changes to this project will be documented in this file.

## [Unreleased]
TBD

## [v2.0.0] - 2025-11-24
There are state-breaking changes in this release.

- Fix: Update Go dependencies to resolve security vulnerabilities.
- Fix: Register missing IBC rate-limiting queries and messages.
- Fix: Resolve issues with the Rate-Limiting ICS4 Wrapper.
- Fix: Prevent the recovery process from using the operator address.
- Doc: Add a deterministic build section to the README documentation.
- Fix: Implement an end-to-end test framework to automate TokenWrapper module testing.
- Fix: Enhance `SendPacket` callback in the TokenWrapper module with additional validations to prevent tokens from being stuck on Axelar.
- Fix: Enable `uzig` tokens to be sent to non-Axelar channels without wrapping.
- Chore: Adopted a versioning scheme where all releases use a `v` prefix, and any governance upgrade triggers a major version increment.
- Chore: Now verify chain version compatibility using the upgrade version specified in `app/upgrades/vXYZ/constants.go`, instead of relying on the binary version.
- Fix: Resolve an issue in the `OnAcknowledgementPacket` callback where acknowledgement deserialization was not handled correctly.
- Chore: Upgrade Cosmos SDK from v0.50.14 to v0.53.4
- Chore: Add `authtypes.ModuleName` to preBlockers array as required by Cosmos SDK v0.53.x
- Note: This upgrade requires a coordinated chain upgrade. All validators must upgrade their binaries simultaneously.
- Chore: Raised the minimum required Go version to v1.25.4
- Fix: The `crisis` module has been removed due to its deprecation in Cosmos SDK v0.53.x
- Test: Eliminate all invariant tests associated with the `crisis` module, since they are no longer compatible with Cosmos SDK v0.53.x

## [v1.2.2] - 2025-09-12
There is no breaking changes in this release.

- Fix: Move the client ID validation to occur after the `OnRecvPacket` call in the IBC stack. In the event of an error, send a successful response accompanied by an error event. This ensures the recipient address can still receive the IBC vouchers and prevents tokens from becoming trapped on Axelar.
- Fix: Add the client ID validation into the `OnAcknowledgement` callback. If an error occurs, bypass the wrapping process and emit an error event.
- Fix: In the `OnAcknowledgement` callback, issue an error if there is an attempt to transfer `unit-zig` IBC vouchers over a channel that is not associated with Axelar.
- Fix: Updated Go dependencies to address security vulnerabilities

## [v1.2.1] - 2025-09-09
There is no breaking changes in this release.

- Fix: Ledger support fixed with the new version of ledger-cosmos-go

## [v1.2.0] - 2025-09-02
There is no breaking changes in this release.

- Feat: Increased the maximum allowed size for wasm uploads, enabling support for larger wasm files.
- Fix: Modified the onrecvpacket callback to ensure IBC vouchers are released to the recipient address in the following cases:
  * The module wallet lacks sufficient ZIG tokens
  * The module is disabled
  * The module's IBC settings do not match the IBC packet
- Feat: Added a new permissionless "recover zig" message, allowing users to reclaim their ZIG tokens if a bridge transfer fails.
- Refactor: Moved IBC handler functions into the keeper for improved code organization.

**MsgRecoverZig:**
```protobuf
// v1.2.0 - NEW
message MsgRecoverZig {
  option (cosmos.msg.v1.signer) = "signer";
  string signer = 1;
  string address = 2;
}
```

**MsgRecoverZigResponse:**
```protobuf
// v1.2.0 - NEW
message MsgRecoverZigResponse {
  string signer = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  string receiving_address = 2
      [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  cosmos.base.v1beta1.Coin locked_ibc_amount = 3
      [ (gogoproto.nullable) = false ];
  cosmos.base.v1beta1.Coin unlocked_native_amount = 4
      [ (gogoproto.nullable) = false ];
}
```

## [v1.1.2] - 2025-08-07
There is no breaking changes in this release.

## [v1.1.1] - 2025-08-06
There is no breaking changes in this release.

- Fix: Denom subdenom validation to allow hyphens
- Fix: Update IBC settings allows to use hyphens in the module denom

## [v1.1.0] - 2025-07-22
This document outlines the breaking changes between zigchain v1.0.0 and v1.1.0.

### Overview
The main changes in v1.1.0 include:
- **Go version requirement**: Minimum required Go version updated to v1.24
- **Field naming convention**: All `camelCase` field names have been converted to `snake_case`
- **New admin management system**: Replaced direct admin updates with a 2-step proposal/claim process
- **Field renames**: Several fields have been renamed for clarity (e.g., `maxSupply` → `minting_cap`)
- **New messages**: Added new functionality for pauser management and operator address management
- **Enhanced responses**: Many response messages now include additional fields
- **New functionality**: Pauser management and improved operator address management
- **Improved IBC settings**: More detailed IBC configuration options

### Factory Module Changes

#### Breaking Changes in Messages

##### 1. Field Naming Convention Changes

All camelCase field names have been converted to snake_case across all messages:

**MsgCreateDenom:**
```protobuf
// v1.0.0
message MsgCreateDenom {
  string subDenom = 2;
  string maxSupply = 3;
  bool canChangeMaxSupply = 4;
  string URIHash = 6;
}

// v1.1.0
message MsgCreateDenom {
  string sub_denom = 2;
  string minting_cap = 3;
  bool can_change_minting_cap = 4;
  string URI_hash = 6;
}
```

**MsgCreateDenomResponse:**
```protobuf
// v1.0.0
message MsgCreateDenomResponse {
  string bankAdmin = 2;
  string metadataAdmin = 3;
  string maxSupply = 5;
  bool canChangeMaxSupply = 6;
  string URIHash = 8;
}

// v1.1.0
message MsgCreateDenomResponse {
  string bank_admin = 2;
  string metadata_admin = 3;
  string minting_cap = 5;
  bool can_change_minting_cap = 6;
  string URI_hash = 8;
}
```

**MsgMintAndSendTokensResponse:**
```protobuf
// v1.0.0
message MsgMintAndSendTokensResponse {
  cosmos.base.v1beta1.Coin tokenMinted = 1;
  cosmos.base.v1beta1.Coin totalMinted = 3;
  cosmos.base.v1beta1.Coin totalSupply = 4;
}

// v1.1.0
message MsgMintAndSendTokensResponse {
  cosmos.base.v1beta1.Coin token_minted = 1;
  cosmos.base.v1beta1.Coin total_minted = 3;
  cosmos.base.v1beta1.Coin total_supply = 4;
}
```

**MsgUpdateDenomURI:**
```protobuf
// v1.0.0
message MsgUpdateDenomURI {
  string URIHash = 4;
}

// v1.1.0
message MsgUpdateDenomURI {
  string URI_hash = 4;
}
```

**MsgUpdateDenomURIResponse:**
```protobuf
// v1.0.0
message MsgUpdateDenomURIResponse {
  string URIHash = 3;
}

// v1.1.0
message MsgUpdateDenomURIResponse {
  string URI_hash = 3;
}
```

**MsgBurnTokensResponse:**
```protobuf
// v1.0.0
message MsgBurnTokensResponse {
  cosmos.base.v1beta1.Coin amountBurned = 1;
}

// v1.1.0
message MsgBurnTokensResponse {
  cosmos.base.v1beta1.Coin amount_burned = 1;
}
```

##### 2. Service Method Changes

**Removed Method:**
```protobuf
// v1.0.0 - REMOVED
rpc UpdateDenomAuth (MsgUpdateDenomAuth) returns (MsgUpdateDenomAuthResponse);
```

**Added Methods (2-step admin management):**
```protobuf
// v1.1.0 - NEW
rpc ProposeDenomAdmin(MsgProposeDenomAdmin) returns (MsgProposeDenomAdminResponse);
rpc ClaimDenomAdmin(MsgClaimDenomAdmin) returns (MsgClaimDenomAdminResponse);
rpc DisableDenomAdmin(MsgDisableDenomAdmin) returns (MsgDisableDenomAdminResponse);
```

**Method Rename:**
```protobuf
// v1.0.0
rpc UpdateDenomMaxSupply (MsgUpdateDenomMaxSupply) returns (MsgUpdateDenomMaxSupplyResponse);

// v1.1.0
rpc UpdateDenomMintingCap(MsgUpdateDenomMintingCap) returns (MsgUpdateDenomMintingCapResponse);
```

##### 3. New Messages

**MsgProposeDenomAdmin:**
```protobuf
// v1.1.0 - NEW
message MsgProposeDenomAdmin {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name) = "zigchain/x/factory/MsgProposeDenomAdmin";

  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string denom = 2;
  string bank_admin = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string metadata_admin = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

message MsgProposeDenomAdminResponse {
  string denom = 1;
  string bank_admin = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string metadata_admin = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

**MsgClaimDenomAdmin:**
```protobuf
// v1.1.0 - NEW
message MsgClaimDenomAdmin {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name) = "zigchain/x/factory/MsgClaimDenomAdmin";

  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string denom = 2;
}

message MsgClaimDenomAdminResponse {
  string denom = 1;
}
```

**MsgDisableDenomAdmin:**
```protobuf
// v1.1.0 - NEW
message MsgDisableDenomAdmin {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name) = "zigchain/x/factory/MsgDisableDenomAdmin";

  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string denom = 2;
}

message MsgDisableDenomAdminResponse {
  string denom = 1;
}
```

##### 4. Removed Messages

**MsgUpdateDenomAuth and MsgUpdateDenomAuthResponse** - These messages have been completely removed and replaced with the new 2-step admin management system.

#### Breaking Changes in Queries

##### 1. Field Naming Convention Changes

**QueryDenomResponse:**
```protobuf
// v1.0.0
message QueryDenomResponse {
  string totalMinted = 2;
  string totalSupply = 3;
  string totalBurned = 4;
  string maxSupply = 5;
  bool canChangeMaxSupply = 6;
  string bankAdmin = 8;
  string metadataAdmin = 9;
}

// v1.1.0
message QueryDenomResponse {
  string total_minted = 2;
  string total_supply = 3;
  string total_burned = 4;
  string minting_cap = 5;
  string max_supply = 6;
  bool can_change_minting_cap = 7;
  string bank_admin = 9;
  string metadata_admin = 10;
}
```

**QueryAllDenomAuthResponse:**
```protobuf
// v1.0.0
message QueryAllDenomAuthResponse {
  repeated DenomAuth denomAuth = 1;
}

// v1.1.0
message QueryAllDenomAuthResponse {
  repeated DenomAuth denom_auth = 1;
}
```

**QueryDenomAuthResponse:**
```protobuf
// v1.0.0
message QueryDenomAuthResponse {
  DenomAuth denomAuth = 1;
}

// v1.1.0
message QueryDenomAuthResponse {
  DenomAuth denom_auth = 1;
}
```

#### Breaking Changes in Denom Types

##### 1. Field Naming Convention Changes

**Denom:**
```protobuf
// v1.0.0
message Denom {
  string maxSupply = 3;
  bool canChangeMaxSupply = 5;
}

// v1.1.0
message Denom {
  string minting_cap = 3;
  bool can_change_minting_cap = 5;
}
```

**DenomResponse:**
```protobuf
// v1.0.0
message DenomResponse {
  string maxSupply = 3;
  string totalMinted = 4;
  bool canChangeMaxSupply = 5;
  string totalBurned = 6;
  string totalSupply = 7;
}

// v1.1.0
message DenomResponse {
  string minting_cap = 3;
  string max_supply = 4;
  string total_minted = 5;
  bool can_change_minting_cap = 6;
  string total_burned = 7;
  string total_supply = 8;
}
```

##### 2. New Legacy Type

**LegacyDenom** - A new message type has been added to maintain backward compatibility:
```protobuf
// v1.1.0 - NEW
message LegacyDenom {
  string creator = 1;
  string denom = 2;
  string maxSupply = 3;
  string minted = 4;
  bool canChangeMaxSupply = 5;
}
```

### DEX Module Changes

#### Breaking Changes in Messages

##### 1. Field Naming Convention Changes

**MsgCreatePoolResponse:**
```protobuf
// v1.0.0
message MsgCreatePoolResponse {
  string poolId = 1;
}

// v1.1.0
message MsgCreatePoolResponse {
  string pool_id = 1;
}
```

**MsgSwapExactIn:**
```protobuf
// v1.0.0
message MsgSwapExactIn {
  string poolId = 3;
  cosmos.base.v1beta1.Coin outgoingMin = 5;
}

// v1.1.0
message MsgSwapExactIn {
  string pool_id = 3;
  cosmos.base.v1beta1.Coin outgoing_min = 5;
}
```

**MsgSwapExactInResponse:**
```protobuf
// v1.0.0
message MsgSwapExactInResponse {
  string poolId = 1;
  cosmos.base.v1beta1.Coin outgoingMin = 6;
}

// v1.1.0
message MsgSwapExactInResponse {
  string pool_id = 1;
  cosmos.base.v1beta1.Coin outgoing_min = 6;
}
```

**MsgSwapExactOut:**
```protobuf
// v1.0.0
message MsgSwapExactOut {
  string poolId = 3;
  cosmos.base.v1beta1.Coin incomingMax = 5;
}

// v1.1.0
message MsgSwapExactOut {
  string pool_id = 3;
  cosmos.base.v1beta1.Coin incoming_max = 5;
}
```

**MsgSwapExactOutResponse:**
```protobuf
// v1.0.0
message MsgSwapExactOutResponse {
  string poolId = 1;
  cosmos.base.v1beta1.Coin incomingMax = 6;
}

// v1.1.0
message MsgSwapExactOutResponse {
  string pool_id = 1;
  cosmos.base.v1beta1.Coin incoming_max = 6;
}
```

**MsgAddLiquidity:**
```protobuf
// v1.0.0
message MsgAddLiquidity {
  string poolId = 2;
}

// v1.1.0
message MsgAddLiquidity {
  string pool_id = 2;
}
```

**MsgAddLiquidityResponse:**
```protobuf
// v1.0.0
message MsgAddLiquidityResponse {
  cosmos.base.v1beta1.Coin lptoken = 1;
}

// v1.1.0
message MsgAddLiquidityResponse {
  cosmos.base.v1beta1.Coin lptoken = 1;
  cosmos.base.v1beta1.Coin actual_base = 2;
  cosmos.base.v1beta1.Coin actual_quote = 3;
  repeated cosmos.base.v1beta1.Coin returned_coins = 4;
}
```

### Tokenwrapper Module Changes

#### Breaking Changes in Messages

##### 1. Field Naming Convention Changes

**MsgFundModuleWalletResponse:**
```protobuf
// v1.0.0
message MsgFundModuleWalletResponse {
  string moduleAddress = 4;
}

// v1.1.0
message MsgFundModuleWalletResponse {
  string module_address = 4;
}
```

**MsgWithdrawFromModuleWalletResponse:**
```protobuf
// v1.0.0
message MsgWithdrawFromModuleWalletResponse {
  string moduleAddress = 4;
}

// v1.1.0
message MsgWithdrawFromModuleWalletResponse {
  string module_address = 4;
}
```

##### 2. Enhanced Response Messages

**MsgUpdateParamsResponse:**
```protobuf
// v1.0.0
message MsgUpdateParamsResponse {}

// v1.1.0
message MsgUpdateParamsResponse {
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  Params params = 2 [(gogoproto.nullable) = false, (amino.dont_omitempty) = true];
}
```

**MsgEnableTokenWrapperResponse:**
```protobuf
// v1.0.0
message MsgEnableTokenWrapperResponse {}

// v1.1.0
message MsgEnableTokenWrapperResponse {
  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  bool enabled = 2;
}
```

**MsgDisableTokenWrapperResponse:**
```protobuf
// v1.0.0
message MsgDisableTokenWrapperResponse {}

// v1.1.0
message MsgDisableTokenWrapperResponse {
  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  bool enabled = 2;
}
```

##### 3. New Service Methods

**Added Methods:**
```protobuf
// v1.1.0 - NEW
rpc AddPauserAddress(MsgAddPauserAddress) returns (MsgAddPauserAddressResponse);
rpc RemovePauserAddress(MsgRemovePauserAddress) returns (MsgRemovePauserAddressResponse);
rpc ProposeOperatorAddress(MsgProposeOperatorAddress) returns (MsgProposeOperatorAddressResponse);
rpc ClaimOperatorAddress(MsgClaimOperatorAddress) returns (MsgClaimOperatorAddressResponse);
```

**Removed Method:**
```protobuf
// v1.0.0 - REMOVED
rpc UpdateOperatorAddress(MsgUpdateOperatorAddress) returns (MsgUpdateOperatorAddressResponse);
```

##### 4. New Messages

**MsgAddPauserAddress:**
```protobuf
// v1.1.0 - NEW
message MsgAddPauserAddress {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name) = "zigchain/x/tokenwrapper/MsgAddPauserAddress";

  string signer = 1;
  string new_pauser = 2;
}

message MsgAddPauserAddressResponse {
  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  repeated string pauser_addresses = 6 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

**MsgRemovePauserAddress:**
```protobuf
// v1.1.0 - NEW
message MsgRemovePauserAddress {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name) = "zigchain/x/tokenwrapper/MsgRemovePauserAddress";

  string signer = 1;
  string pauser = 2;
}

message MsgRemovePauserAddressResponse {
  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  repeated string pauser_addresses = 6 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

**MsgProposeOperatorAddress:**
```protobuf
// v1.1.0 - NEW
message MsgProposeOperatorAddress {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name) = "zigchain/x/tokenwrapper/MsgProposeOperatorAddress";

  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string new_operator = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

message MsgProposeOperatorAddressResponse {
  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string proposed_operator_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

**MsgClaimOperatorAddress:**
```protobuf
// v1.1.0 - NEW
message MsgClaimOperatorAddress {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name) = "zigchain/x/tokenwrapper/MsgClaimOperatorAddress";

  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

message MsgClaimOperatorAddressResponse {
  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string operator_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

##### 5. Enhanced IBC Settings

**MsgUpdateIbcSettings:**
```protobuf
// v1.0.0
message MsgUpdateIbcSettings {
  string client_id = 2;
  string source_port = 3;
  string source_channel = 4;
  string denom = 5;
  uint32 decimal_difference = 6;
}

// v1.1.0
message MsgUpdateIbcSettings {
  string native_client_id = 2;
  string counterparty_client_id = 3;
  string native_port = 4;
  string counterparty_port = 5;
  string native_channel = 6;
  string counterparty_channel = 7;
  string denom = 8;
  uint32 decimal_difference = 9;
}
```

**MsgUpdateIbcSettingsResponse:**
```protobuf
// v1.0.0
message MsgUpdateIbcSettingsResponse {}

// v1.1.0
message MsgUpdateIbcSettingsResponse {
  string signer = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string native_client_id = 2;
  string counterparty_client_id = 3;
  string native_port = 4;
  string counterparty_port = 5;
  string native_channel = 6;
  string counterparty_channel = 7;
  string denom = 8;
  uint32 decimal_difference = 9;
}
```

##### 6. Removed Messages

**MsgUpdateOperatorAddress and MsgUpdateOperatorAddressResponse** - These messages have been completely removed and replaced with the new 2-step operator address management system.

### Migration Guide

#### For Factory Module

1. **Update field names**: Convert all camelCase field names to snake_case
2. **Replace direct admin updates**: Use the new 2-step process:
   - First call `ProposeDenomAdmin` to propose new admins
   - Then call `ClaimDenomAdmin` to claim the admin role
3. **Update method calls**: Replace `UpdateDenomMaxSupply` with `UpdateDenomMintingCap`

#### For DEX Module

1. **Update field names**: Convert all camelCase field names to snake_case
2. **Handle new response fields**: The `MsgAddLiquidityResponse` now includes `returned_coins`

#### For Tokenwrapper Module

1. **Update field names**: Convert all camelCase field names to snake_case
2. **Replace direct operator updates**: Use the new 2-step process:
   - First call `ProposeOperatorAddress` to propose a new operator
   - Then call `ClaimOperatorAddress` to claim the operator role
3. **Use new pauser management**: Use `AddPauserAddress` and `RemovePauserAddress` for pauser management
4. **Update IBC settings**: The IBC settings now require separate native and counterparty configurations
