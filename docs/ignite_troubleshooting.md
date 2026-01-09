# Ignite Troubleshooting 

## Error with Bank versions when scaffolding new module with ignite

When scaffolding a new module with ignite, you get the following error:
```sh
➜ ignite scaffold module bluemod

✘ Error while running command go mod tidy: go: finding module for package cosmossdk.io/x/bank/types

go: found cosmossdk.io/x/bank/types in cosmossdk.io/x/bank v0.2.0-rc.1

go: github.com/cosmos/cosmos-sdk@v0.52.0: reading github.com/cosmos/cosmos-sdk/go.mod at revision v0.52.0: unknown revision v0.52.0

: exit status 1
```

This is because the version of the bank module is not compatible with the version of the cosmos-sdk. To change the version of the bank module is not possible as during the proto build process it gets overwritten. While the error appears, this happens after the files scaffolding has been completed. Making it possible just to recover those previous files to continue with the development.

Steps are:
```sh
# To create new module "bluemod"
ignite scaffold module bluemod

# It will return bank error due to version but create the files
# Run this now:
chmod +x sh/revert-buf.sh
sh/revert-buf.sh

# You should be able to serve the blockchain successfully now.
ignite chain serve --reset-once

# run on parallel terminal to ensure that works as expected
zigchaind q bluemod --help
```

If you want to revert the changes, as tons of files are created, use:
```sh
git reset --hard HEAD; git clean -fd
```