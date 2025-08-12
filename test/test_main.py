
import os, shutil, subprocess, time, gzip, sys, logging, threading
import pytest, asyncio, json, signal, requests

from compare_results import check_blockupdates_values, check_cardinal_values

logging.basicConfig(level=logging.INFO, format="%(levelname)s - %(message)s")

DATADIR = './resources/datadir/'
rpc_url = "http://127.0.0.1:8545"
geth = None

def import_chain():
        logging.info("importing chain")
        try:
            rpc = {"jsonrpc":"2.0", "method":"admin_importChain", "params":["./resources/midChain.gz"], "id":22}
            response = requests.post(rpc_url, json=rpc).json()
            if "error" in response:
                logging.warning(f"{response['error']['message']}")

                time.sleep(5) 
                b = get_block_number()
                if not (b and b > 0):
                    logging.error("Chain import may have failed, latest block not found")
                    sys.exit(1)
        except Exception as e:
            logging.error(f"Failed to import chain: {e}")
            raise
        
def decompress_control_data():
    logging.info("decompressing control cardinal data")
    with gzip.open('./resources/v1.14.7.0.5-control-data/p1cs.json.gz', "rb") as f:
        with open('./resources/control_card_data.json', "wb") as f_o:
            shutil.copyfileobj(f, f_o)

    logging.info("decompressing control plugeth data")
    with gzip.open('./resources/v1.14.7.0.5-control-data/p1bu.json.gz', "rb") as f:
        with open('./resources/control_plugeth_data.json', "wb") as f_o:
            shutil.copyfileobj(f, f_o)

def cleanup():
    logging.info("cleanup")

    files_to_remove = [
        './resources/test_card_data.json',
        './resources/test_plugeth_data.json',
        './resources/control_card_data.json',
        './resources/control_plugeth_data.json'
    ]

    for path in files_to_remove:
        if os.path.exists(path):
            os.remove(path)

    if os.path.exists(DATADIR):
        shutil.rmtree(DATADIR)

def get_block_number():
    time.sleep(5)
    rpc = {"jsonrpc":"2.0", "method":"eth_blockNumber", "params":[], "id":1}
    try:
        response = requests.post(rpc_url, json=rpc).json()
        return int( response['result'], 16)
    except Exception as e:
        logging.error(f"error in getting block no: {e}")
        return None

def start_node(path_to_bin):
    logging.info("starting geth")
    if not os.path.exists(DATADIR):
       os.makedirs(DATADIR)

    global geth 
    geth = subprocess.Popen([
        f"{path_to_bin}",
        "--nodiscover",
        "--holesky",
        "--http",
        "--http.api=eth,admin,plugeth,cardinal",
        "--ws",
        "--ws.api=cardinal,plugeth",
        "--verbosity=0",
        f"--datadir={DATADIR}"])

    time.sleep(5)

    try:
        subprocess.Popen(["python3", "ws_data_capture.py", "test_plugeth_data", "plugeth"])
        time.sleep(2)
        import_chain()
        time.sleep(15)
        subprocess.Popen(["python3", "ws_data_capture.py", "test_card_data", "cardinal"])

    except Exception as e:
        logging.error(f"An error occurred: {e}")
        terminate_geth()
        raise
        sys.exit(1)

def terminate_geth():
    if geth:
        geth.terminate()
        geth.wait()

    
def monitor_node():
    time.sleep(10)
               
    while True:
        blockno = get_block_number()
        if blockno and blockno >= 2000:
            logging.info(f"block number {blockno} reached, stopping node")
            time.sleep(30)
            terminate_geth()
            time.sleep(5)
            break
        time.sleep(7)

def gather_data(binary_path):
    logging.info("Gathering data")
    node_thread = threading.Thread(target=start_node, args=(binary_path,))
    monitor_thread = threading.Thread(target=monitor_node)
    
    node_thread.start()
    monitor_thread.start()

    node_thread.join()
    monitor_thread.join()

def test_main(bin_path):

    gather_data(bin_path)

    decompress_control_data()

    check_blockupdates_values()

    check_cardinal_values()
    
    cleanup()

# NOTE the binary produced by the build tool will not work for this test. The xplugeth_imports.go in workdir/cmd/geth will need to be modified to point to 
# xplugeth/plugins/merge and, after running go get, the binary will need to be built with the -tags=xplugeth tag. 

# pytest --bin-path=/path/to/binary if being run with pytest 
# python3 test_main.py /path/to/binary for debugging


if __name__ == '__main__':
   test_main(sys.argv[1])