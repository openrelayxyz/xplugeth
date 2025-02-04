import os, shutil, subprocess, time, gzip, sys, logging
from compare_cardinal import test_cardinal
from ws_data_capture import subscribe_to_websocket

logging.basicConfig(level=logging.INFO, format="%(levelname)s - %(message)s")

DATADIR = './resources/datadir/'
XPLUGETH_PATH = '/Users/jesseakoh/Desktop/work/code/OpenRelay/xplugeth'


def import_chain():
    logging.info("importing chain")
    import_command = (
        "curl 127.0.0.1:8545 "
        "-H 'Content-Type: application/json' "
        "--data '{\"jsonrpc\": \"2.0\", \"method\": \"admin_importChain\", \"params\": [\"./resources/midChain.gz\"], \"id\": 22}'"
    )
    result = subprocess.run(import_command, shell=True)
    if result.returncode != 0 :
        logging.error("Chain import failed: Unable to connect to 127.0.0.1:8545")
        sys.exit(1)

def decompress_control_data():
    logging.info("decompressing control data")
    with gzip.open('./resources/control_card_data.json.gz', "rb") as f:
        with open('./resources/control_card_data.json', "wb") as f_o:
            shutil.copyfileobj(f, f_o)

def cleanup():
    logging.info("cleanup")
    if os.path.exists("./resources/test_card_data.json"):
        os.remove("./resources/test_card_data.json")
    
    with open("./resources/control_card_data.json", "rb") as f:
        with gzip.open('./resources/control_card_data.json.gz', "wb") as f_o:
            shutil.copyfileobj(f, f_o)
  
def main():
    logging.info("building geth")
    build_path = os.path.abspath('../build/build.py')
    build_command = (
        f"python3 {build_path} "
        "-s https://github.com/ethereum/go-ethereum "
        "-p github.com/openrelayxyz/xplugeth/plugins/merge@v0.12.0 "
        f"-r github.com/openrelayxyz/xplugeth={XPLUGETH_PATH}"
    )
    print(build_command)
    subprocess.run(build_command, shell=True)

    if os.path.exists(DATADIR):
        shutil.rmtree(DATADIR)
    os.makedirs(DATADIR)
    if not os.path.exists("./resources/geth"):
        shutil.copy("/tmp/output/geth", "./resources/geth")
        
    print(">starting the node")    
    process = subprocess.Popen(
        f"./resources/geth --nodiscover --holesky "
        "--http --http.api=eth,admin,plugeth,cardinal "
        "--ws --ws.api=cardinal,plugeth "
        f"--datadir={DATADIR}",
        shell=True,
    )
    time.sleep(5)

    try:
        import_chain()
        subscribe_to_websocket('test_card_data', 'cardinal')
    except Exception as e:
        logging.error(f"An error occurred: {e}")
        sys.exit(1)
    finally:
        time.sleep(2)
        process.terminate()
        process.wait()
    
    decompress_control_data()
    test_cardinal()
    cleanup()


if __name__ == '__main__':
    main()