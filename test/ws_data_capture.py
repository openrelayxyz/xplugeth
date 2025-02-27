import asyncio, logging, websockets, json, sys

logging.basicConfig(level=logging.INFO, format="%(levelname)s - %(message)s")

async def subscribe_to_websocket(file_name, namespace):
    output = []
    output_file = file_name
    ws_url = "ws://localhost:8546"
    if namespace == 'cardinal':
        ws_url = "ws://localhost:8555"

    try:
        async with websockets.connect(ws_url) as websocket:
            if namespace == "cardinal":
                logging.info("gathering cardinal streams data")
                for i in range(0, 2000):
                    req = json.dumps({
                        "jsonrpc": "2.0",
                        "method": "cardinal_streamsBlock",
                        "params": [f'{hex(i)}'],
                        "id": 1
                    })
                    await websocket.send(req)
            else:
                logging.info("gathering plugeth blockupdates data")
                req = json.dumps({
                    "jsonrpc": "2.0",
                    "method": "plugeth_subscribe",
                    "params": ["blockUpdates"],
                    "id": 1
                })
                await websocket.send(req)


            while True:
                try:
                    response = await websocket.recv()
                    output.append(json.loads(response))

                    if len(output) >= 1999:
                        logging.info(f"writing to output file {output_file}.json")
                        with open(f'./resources/{output_file}.json', 'w') as f:
                            json.dump(output, f)
                        await websocket.close()
                        logging.info("WebSocket connection closed.")
                        break
                    
                except websockets.ConnectionClosed as e:
                    logging.error(f"Websocket closed unexpectedly: {e}")
                    break
                except asyncio.CancelledError:
                    logging.error("WebSocket subscription was cancelled.")
                    break
                except Exception as e:
                    logging.error(f"Unexpected error: {e}")
                    break
    except Exception as e:
        logging.Error(f"failed to connect to websocket {e}")
        
if __name__ == "__main__":
    asyncio.run(subscribe_to_websocket(sys.argv[1], sys.argv[2]))
