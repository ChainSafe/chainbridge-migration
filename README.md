# chainbridge-migration
Helper script for migrating ChainBridge from v1 to v2

## Description
The idea of this script is to ease the process of pausing bridge contracts while migrating from v1 to v2 of ChainBridge.
It will primarily check if all proposals have been resolved (for all chains being bridged by v1 of ChainBridge) and then pause bridge contract for each chain (only if `autoPauseBridge` configuration property is set to `true`)

The script goes through all `ProposalEvents` emitted by [bridge contract](https://github.com/ChainSafe/chainbridge-solidity/blob/release/v1.0.0/contracts/Bridge.sol#L57) and parses if there are any Proposals that haven't been resolved (meaning Proposals with statuses _Active_ or _Passed_).
This process is being executed for each chain defined in v1 ChainBridge configuration.
All pending Proposals are displayed inside the console with some additional details.

Described check for all pending Proposals will be restarted every 60 seconds until all pending Proposals have been resolved.
After all pending Proposals are resolved, if `autoPauseBridge` configuration property is set to `true`, script will execute [`adminPauseTransfers`](https://github.com/ChainSafe/chainbridge-solidity/blob/release/v1.0.0/contracts/Bridge.sol#L147) on each bridge contract.

## How to use it

### 1) Clone repo
Clone this repo to your local machine.

### 2) Create configuration
Define a **chainbridge-migration** configuration file. The script expects this file to be defined as `configuration.json` in the root of the project. See [Configuration](#configuration) for more details.

### 3) Start script
Once the configuration has been created, the script can be started by running `make stop-bridge`.


## Configuration

- `configurationPath` - **[_required_]** - path to v1 ChainBridge configuration file.
- `startingBlocks` - **[_optional_]** - mapping of **chain ID**** <> **starting block**. Defines from which block should script process events for each chain. If starting block for one chain is omitted (or this property is entirely omitted) script will start querying from the first block.
- `autoPauseBridge` - **[_optional_]** - boolean value that defines if script should automatically execute [`adminPauseTransfers`](https://github.com/ChainSafe/chainbridge-solidity/blob/release/v1.0.0/contracts/Bridge.sol#L147) on each bridge contract after all Proposals are _Executed_ or _Cancelled_. 
- `privateKeys` - **[_required if `autoPauseBridge` set to true_]** - mapping of **chain ID**** <> **private key**. Defines administrator private keys for each bridge contract, used to execute pausing bridge.

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
  "autoPauseBridge": false
}
```