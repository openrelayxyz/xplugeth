import os, shutil, subprocess, time, gzip
from compare_cardinal import test_cardinal
from ws_data_capture import subscribe_to_websocket

DATADIR = './datadir/'
XPLUGETH_PATH = '/Users/jesseakoh/Desktop/work/code/OpenRelay/xplugeth'


def import_chain():
    print(">importing chain")
    import_command = (
        "curl 127.0.0.1:8545 "
        "-H 'Content-Type: application/json' "
        "--data '{\"jsonrpc\": \"2.0\", \"method\": \"admin_importChain\", \"params\": [\"./resources/midChain.gz\"], \"id\": 22}'"
    )
    subprocess.run(import_command, shell=True)

def decompress_control_data():
    with gzip.open('./resources/control_card_data.json.gz', "rb") as f:
        with open('./resources/control_card_data.json', "wb") as f_o:
            shutil.copyfileobj(f, f_o)

def cleanup():
    if os.path.exists("./resources/test_card_data.json"):
        os.remove("./resources/test_card_data.json")
    
    with open("./resources/control_card_data.json", "rb") as f:
        with gzip.open('./resources/control_card_data.json.gz', "wb") as f_o:
            shutil.copyfileobj(f, f_o)
  
def main():
    print(">building geth")
    build_path = os.path.abspath('../build/build.py')
    build_command = (
        f"python3 {build_path} "
        "-s https://github.com/ethereum/go-ethereum "
        "-p github.com/openrelayxyz/xplugeth/plugins/merge@v0.12.0 "
        f"-r github.com/openrelayxyz/xplugeth={XPLUGETH_PATH} "
        "-a ./resources/"
    )
    print(build_command)
    subprocess.run(build_command, shell=True)

    if os.path.exists(DATADIR):
        shutil.rmtree(DATADIR)
    os.makedirs(DATADIR)

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
        subscribe_to_websocket('test_card_data',  'cardinal')
    finally: 
        time.sleep(2)
        process.terminate()
        process.wait()
    
    decompress_control_data()
    test_cardinal()
    cleanup()


if __name__ == '__main__':
    main()