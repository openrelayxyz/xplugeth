
import os, shutil, subprocess, time, gzip, sys, logging, threading
import pytest, asyncio, json, signal, requests

from compare_cardinal import test_cardinal

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
            sys.exit(1)
        
def decompress_control_data():
    logging.info("decompressing control data")
    with gzip.open('./resources/v1.14.7.0.5-control-data/p1cs.json.gz', "rb") as f:
        with open('./resources/control_card_data.json', "wb") as f_o:
            shutil.copyfileobj(f, f_o)

    with gzip.open('./resources/v1.14.7.0.5-control-data/p1bu.json.gz', "rb") as f:
        with open('./resources/control_plugeth_data.json', "wb") as f_o:
            shutil.copyfileobj(f, f_o)

def cleanup():
    logging.info("cleanup")

    files_to_remove = [
        './resources/test_card_data.json',
        './resources/test_plugeth_data.json',
        # './resources/geth', 
        './resources/control_card_data.json',
        './resources/control_plugeth_data.json'
    ]

    for path in files_to_remove:
        if os.path.exists(path):
            os.remove(path)

    if os.path.exists(DATADIR):
        shutil.rmtree(DATADIR)
    
  
def build():
    logging.info("building geth")
    build_path = os.path.abspath('../build/build.py')
    build_command = (
        f"python3 {build_path} "
        "-s https://github.com/ethereum/go-ethereum "
        "-p github.com/openrelayxyz/xplugeth/plugins/merge@v0.12.0 "
        f"-r github.com/openrelayxyz/xplugeth={os.path.abspath('../')} "
        f"--artifacts-directory={os.path.abspath('./resources')}"
    )
    print(build_command)
    result = subprocess.run(build_command, shell=True)
    if result.returncode != 0:
        logging.error("build failed")
        sys.exit(1)

def get_block_number():
    time.sleep(5)
    rpc = {"jsonrpc":"2.0", "method":"eth_blockNumber", "params":[], "id":1}
    try:
        response = requests.post(rpc_url, json=rpc).json()
        return int( response['result'], 16)
    except Exception as e:
        logging.error(f"error in getting block no: {e}")
        return None

def start_node():
    if not os.path.exists(DATADIR):
       os.makedirs(DATADIR)

    print(">starting the node")   
    # for the sake of macOs issues (Sequioa 15.0 or below) in running binaries with partial or invalid signatures i'll need to have this here 
    # subprocess.run(["codesign", "--force", "--deep", "--sign",  "-", "./resources/geth"])  

    global geth 
    geth = subprocess.Popen(
        f"./resources/geth --nodiscover --holesky "
        "--http --http.api=eth,admin,plugeth,cardinal "
        "--ws --ws.api=cardinal,plugeth "
        "--verbosity=0 "
        f"--datadir={DATADIR}",
        shell=True,
    )

    time.sleep(5)

    try:
        subprocess.Popen(["python3", "ws_data_capture.py", "test_plugeth_data", "plugeth"])
        time.sleep(2)
        import_chain()
        time.sleep(2)
        subprocess.Popen(["python3", "ws_data_capture.py", "test_card_data", "cardinal"])

    except Exception as e:
        logging.error(f"An error occurred: {e}")
        sys.exit(1)

    return geth

    
def monitor_node():
    time.sleep(10)
               
    while True:
        blockno = get_block_number()
        if blockno and blockno >= 2000:
            logging.info(f"block number {blockno} reached, stopping node")
            if geth:
                geth.send_signal(signal.SIGINT)
            time.sleep(5)
            break
        time.sleep(7)
            
def run_test():
    logging.info("running test")
    decompress_control_data()
    test_cardinal()  
    pytest.main(["-q", "--disable-warnings"])  
    
def main():
    build()

    node_thread = threading.Thread(target=monitor_node)
    
    monitor_thread = threading.Thread(target=start_node)

    node_thread.start()
    monitor_thread.start()

    node_thread.join()
    monitor_thread.join()

    run_test()
    cleanup()


if __name__ == '__main__':
   main()