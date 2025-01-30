# this branch has an update to the build.py which enables replacing of the xplugeth import in the go.mod with a local
# version of the project. 
# builds are obtained by running:
# python3 build.py -s https://github.com/ethereum/go-ethereum -p github.com/openrelayxyz/xplugeth/plugins/merge@v0.12.0 -r github.com/openrelayxyz/xplugeth=/path/to/your/ocal/xplugeth 

# once the binary is acquired turn on the node:
$GETH --nodiscover --holesky --http --http.api=eth,admin,plugeth,cardinal --ws --ws.api=cardinal,plugeth

#upload the chain segment. from within the resources directory in xplugeth/test/resources run:
curl 127.0.0.1:8545 -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"admin_importChain","params":["./midChain.gz"],"id":22}' | jq

#wait until the chain has finised and while it is still running run from within test/:
python3 ws_data_capture.py test_card_data cardinal

# at this point you can kill the node and run from within test/:
gunzip ./resources/control_card_data.json.gz
python3 compare_cardianl.py

# the test should finish without a problem at which point test_card_data.json can be deleted and the control data re-zipped.