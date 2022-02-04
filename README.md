# chainbridge-migration
_Helper scripts for migrating ChainBridge from v1 to v2_

### For full migration guide see our [wiki](https://github.com/ChainSafe/chainbridge-migration/wiki)

## Description
The idea of these scripts is to ease the process of migrating from v1 to v2 of ChainBridge.
You can execute two different scripts `stop-bridge` and `transfer-tokens`.

### `stop-bridge`
The script will primarily check if all proposals have been resolved (for all chains defined in the configuration of v1 of ChainBridge) and then pause bridge contract for each chain (only if `autoPauseBridge` configuration property is set to `true`)

The script goes through all `ProposalEvents` emitted by [bridge contract](https://github.com/ChainSafe/chainbridge-solidity/blob/release/v1.0.0/contracts/Bridge.sol#L57) and parses if there are any Proposals that haven't been resolved (meaning Proposals with statuses _Active_ or _Passed_).
This process is being executed for each chain defined in v1 ChainBridge configuration.
All pending Proposals are displayed inside the console with some additional details.

The script will restart described check for all pending Proposals every 60 seconds until all pending Proposals have been resolved.
After all pending Proposals are resolved, if `autoPauseBridge` configuration property is set to `true`, script will execute [`adminPauseTransfers`](https://github.com/ChainSafe/chainbridge-solidity/blob/release/v1.0.0/contracts/Bridge.sol#L147) on each bridge contract.

### `transfer-tokens`

The script will go through all tokens defined in the configuration, and execute [`adminWithdraw`](https://github.com/ChainSafe/chainbridge-solidity/blob/release/v1.0.0/contracts/Bridge.sol#L274) on the appropriate bridge contract.

This script is used to ease up migrating liquidity for tokens that are locked/released by handlers.
The destination address defined in the configuration for each token should be set to the appropriate v2 handler so that withdrawal and migration are executed in one transaction.

## How to use it

### 1) Clone repo
Clone this repo to your local machine.

### 2) Create configuration
Define a **chainbridge-migration** configuration file. The script expects this file to be defined as `configuration.json` in the project's root. See [Configuration](#configuration) for more details.

### 3) Start script
Once the configuration has been created, you can start the script by running `make stop-bridge` or `make transfer-tokens`.

## Configuration

- `configurationPath` - **[_required_]** - path to v1 ChainBridge configuration file.
- `startingBlocks` - **[_optional_]** - mapping of **chain ID**** <> **starting block**. Defines from which block should script process events for each chain. If starting block for one chain is omitted (or this property is entirely omitted) script will start querying from the first block.
- `autoPauseBridge` - **[_optional_]** - boolean value that defines if script should automatically execute [`adminPauseTransfers`](https://github.com/ChainSafe/chainbridge-solidity/blob/release/v1.0.0/contracts/Bridge.sol#L147) on each bridge contract after all Proposals are _Executed_ or _Cancelled_. 
- `privateKeys` - **[_required if `autoPauseBridge` set to true_]** - mapping of **chain ID**** <> **private key**. Defines administrator private keys for each bridge contract, used to execute pausing bridge.
- `tokens` - **[_required for executing `transfer-tokens` script_]** - mapping of **chain ID**** <> **array of token descriptor object**. Defines tokens that should be transferred on each chain, each token entry is defined with: _handlerAddress_, _tokenAddress_, _recipient_, _amountOrTokenID_, _type [erc20/erc721]_ 

** _**chain ID** references ID defined inside v1 ChainBridge configuration file_

Below you can see an example of the configuration file:

```json
{
  "configurationPath": "/../../chainbridge-v1/config.json",
  "privateKeys": {
    "0": "f03....3714",
    "1": "889....6dc6"
  },
  "startingBlocks": {
    "0": "6200000",
    "1": "10087009"
  },
  "autoPauseBridge": false,
  "tokens": {
    "0": [
      {
        "handlerAddress": "0x8B99A045FdA384546D391222258a7b4145d96732",
        "tokenAddress": "0xaFF4481D10270F50f203E0763e2597776068CBc5",
        "recipient": "0xff9f4a4Fc82A803bD00052Ed5b90366c8cDa622b",
        "amountOrTokenID": "100",
        "type": "erc20"
      }
    ],
    "1": [
      {
        "handlerAddress": "0xeC7aBE70B7997852E2D713014B75c4Ff4903D3e5",
        "tokenAddress": "0xDF9D74b9f74C9E09bB01308E405718df46FACeDA",
        "recipient": "0x989264b9448206AE1157B6A86f7A6C3f7F3F48A2",
        "amountOrTokenID": "7",
        "type": "erc721"
      }
    ]
  },
}
```