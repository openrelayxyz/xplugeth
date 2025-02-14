import os, shutil, subprocess, time, gzip, sys, logging
# new imports
import threading, time, signal, requests

DATADIR = './resources/datadir/'

geth = None

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

def getBlockNumber():
    time.sleep(5)
    rpc = {"jsonrpc":"2.0","id":0,"method":"eth_blockNumber","params":[]}
    n = requests.post(url='http://localhost:8545',json=rpc).json()['result']
    return int(n,16)

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

def node():
    if not os.path.exists(DATADIR):
        os.makedirs(DATADIR)

    print(">starting the node")   

    global geth
    geth = subprocess.Popen(
        [
            "./resources/geth",
            "--holesky",
            "--nodiscover",
            "--verbosity=0",
            "--http",
            "--http.api=eth,admin,plugeth,cardinal",
            "--ws",
            "--ws.api=cardinal,plugeth",
            f"--datadir={DATADIR}",
        ]
    )
    time.sleep(5)

    import_chain()

    return geth

def monitor_node():
    while True:
        b = getBlockNumber()
        print(f"printing from node monitor, blocknumber {b}")
        if b >= 2000:
            geth.send_signal(signal.SIGINT)
            time.sleep(5)
            if os.path.exists("./resources/geth"):
                os.remove("./resources/geth")
            if os.path.exists(DATADIR):
                shutil.rmtree(DATADIR)
            return

if __name__ == "__main__":
    build()

    thread1 = threading.Thread(target=node)
    thread2 = threading.Thread(target=monitor_node)

    thread1.start()
    thread2.start()