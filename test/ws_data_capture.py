import asyncio
import websockets
import json
import sys


async def subscribe_to_websocket(file_name, namespace):
    output = []
    output_file = file_name
    ws_url = "ws://localhost:8546"
    if namespace == 'cardinal':
        ws_url = "ws://localhost:8555"

    try:
        async with websockets.connect(ws_url) as websocket:
            if namespace == "cardinal":
                for i in range(0, 2000):
                    req = json.dumps({
                        "jsonrpc": "2.0",
                        "method": "cardinal_streamsBlock",
                        "params": [f'{hex(i)}'],
                        "id": 1
                    })
                    await websocket.send(req)
            else:
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
                        print(f"writing to output file {output_file}.json")
                        with open(f'./resources/{output_file}.json', 'w') as f:
                            json.dump(output, f)
                        await websocket.close()
                        print("WebSocket connection closed.")
                        break
                    
                except websockets.ConnectionClosed as e:
                    print(f"Websocket closed unexpectedly: {e}")
                    break
                except asyncio.CancelledError:
                    print("WebSocket subscription was cancelled.")
                    break
                except Exception as e:
                    print(f"Unexpected error: {e}")
                    break
    except Exception as e:
        print(f"failed to connect to websocket {e}")
        

if __name__ == "__main__":
    asyncio.run(subscribe_to_websocket(sys.argv[1], sys.argv[2]))
