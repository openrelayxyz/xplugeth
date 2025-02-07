import os, shutil, subprocess, time, gzip, sys, logging
import pytest, asyncio
from compare_cardinal import test_cardinal
from ws_data_capture import subscribe_to_websocket

logging.basicConfig(level=logging.INFO, format="%(levelname)s - %(message)s")

DATADIR = './resources/datadir/'


def import_chain():
        logging.info("importing chain")
        import_command = (
            "curl 127.0.0.1:8545 "
            "-H 'Content-Type: application/json' "
            "--data '{\"jsonrpc\": \"2.0\", \"method\": \"admin_importChain\", \"params\": [\"./resources/midChain.gz\"], \"id\": 22}'"
        )
        result = subprocess.run(import_command, shell=True)
        if result.returncode != 0 :
            logging.error(" hain import failed: unable to connect to 127.0.0.1:8545")
            sys.exit(1)

def decompress_control_data():
    logging.info("decompressing control data")
    with gzip.open('./resources/v1.14.7.0.5-control-data/p1bu.json.gz', "rb") as f:
    #with gzip.open('./resources/control_card_data.json.gz', "rb") as f:
        with open('./resources/control_card_data.json', "wb") as f_o:
            shutil.copyfileobj(f, f_o)

def cleanup():
    logging.info("cleanup")
    if os.path.exists("./resources/test_card_data.json"):
        os.remove("./resources/test_card_data.json")
    
    if os.path.exists("./resources/geth"):
        os.remove("./resources/geth")

    if os.path.exists(DATADIR):
        shutil.rmtree(DATADIR)
    
    with open("./resources/control_card_data.json", "r") as f:
        with gzip.open('./resources/control_card_data.json.gz', "wb") as f_o:
            shutil.copyfileobj(f, f_o)
  
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

async def start_node():
    if not os.path.exists(DATADIR):
       os.makedirs(DATADIR)

    print(">starting the node")   
    # for the sake of macOs issues in running binaries with partial or invalid signatures i'll need to have this here 
    subprocess.run(["codesign", "--force", "--deep", "--sign",  "-", "./resources/geth"])  
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
        await subscribe_to_websocket('test_card_data', 'plugeth')
    except Exception as e:
        logging.error(f"An error occurred: {e}")
        sys.exit(1)
    finally:
        time.sleep(2)
        process.terminate()
        process.wait()
    
def run_test():
    logging.info("running test")
    decompress_control_data()
    test_cardinal()  
    pytest.main(["-q", "--disable-warnings"])  
    
def main():
    build()
    asyncio.run(start_node())
    run_test()
    cleanup()


if __name__ == '__main__':
    main()
