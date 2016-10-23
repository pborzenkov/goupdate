## goupdate

*goupdate* is a tool to update Go built binaries installed by `go get`.

## Usage

To update all Go built binaries in $GOPATH/bin:

    $ goupdate [-force] [-verbose]

Use `-force` to update without asking for confirmation.

To update specific binaries:

    $ goupdate <bin1> [bin2]...
