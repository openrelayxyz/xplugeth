package hooktest

import (
	"fmt"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/openrelayxyz/xplugeth/utils"
)

func copyTestResources() error {
	chainid, ok := utils.GetChainID() 
	if !ok {
		panic(fmt.Sprintf("could not resolve chain id from xplugeth utils, hooktest"))
	}
	var network string
	switch chainid {
	case 17000:
		network = "foundation"
	case 137:
		network = "bor"
	default:
		panic(fmt.Sprintf("did not recognize chain id, hooktest"))	
	}

	log.Error("this is the network", "network", network)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current file path")
	}

	packageDir := filepath.Dir(filename)

	sourceDir := filepath.Join(packageDir, network + "-test")
	destDir := "./test/testDataDir"

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext((file.Name())) == ".gz" {
			sourcePath := filepath.Join(sourceDir, file.Name())
			destPath := filepath.Join(destDir, file.Name())

			sourceData, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to read source file %s: %w", file.Name(), err)
			}

			err = os.WriteFile(destPath, sourceData, 0644)
			if err != nil {
				return fmt.Errorf("failed to write destination file %s: %w", file.Name(), err)
			}
		}
	}
	log.Info("Test resources copied", "from", sourceDir, "to", destDir)
	return nil
}

func getBlockNumber() (string, error) {
	var num string
	if err := client.Call(&num, "eth_blockNumber"); err != nil {
		return "", err
	} else {
		return num, nil
	}
}

func coreControlDataDecompress() (map[uint64]map[string]interface{}, error) {
	file, err := os.ReadFile("./test/testDataDir/core-control.json.gz")
	if err != nil {
		log.Error("cannot read file control.json.gz")
		return nil, err
	}
	r, err := gzip.NewReader(bytes.NewReader(file))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	raw, err := io.ReadAll(r)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, err
	}

	var newheadObj map[uint64]map[string]interface{}
	json.Unmarshal(raw, &newheadObj)
	return newheadObj, nil
}

func stateControlDataDecompress() (map[uint64]map[string]interface{}, error) {
	file, err := os.ReadFile("./test/testDataDir/state-control.json.gz")
	if err != nil {
		log.Error("cannot read file control.json.gz")
		return nil, err
	}
	r, err := gzip.NewReader(bytes.NewReader(file))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	raw, err := io.ReadAll(r)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, err
	}

	var stateObject map[uint64]map[string]interface{}
	json.Unmarshal(raw, &stateObject)
	return stateObject, nil
}

var httpClient           = &http.Client{Transport: &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConnsPerHost:   16,
	MaxIdleConns:          16,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}}

func callRPC(method string, params interface{}) error {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := httpClient.Post("http://localhost:8545", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Info("RPC response plugin test", "method", method, "response", string(body))
	return nil
}