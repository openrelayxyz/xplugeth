module github.com/openrelayxyz/xplugeth

go 1.22

// In order to avoid indirect imports which cause conflicts across networks never run go mod tidy on this project.

require (
    github.com/Shopify/sarama v1.38.1
)