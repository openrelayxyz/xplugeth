module github.com/openrelayxyz/xplugeth

go 1.22

require github.com/ethereum/go-ethereum v1.14.6

// In order to avoid indirect imports which cause conflicts across networks never run go mod tidy on this project.
