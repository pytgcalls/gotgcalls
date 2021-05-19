import asyncio
import json
import websockets
from websockets.exceptions import ConnectionClosedError


class StartWebApp:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def _start_web_app(self, limit: int = 10):
        await asyncio.sleep(0.125)
        uri = f'ws://localhost:{self.pytgcalls._port}/go_socket'
        try:
            async with websockets.connect(uri) as websocket:
                self.pytgcalls._gws = websocket
                while True:
                    result_message = await websocket.recv()
                    if result_message == 'PING':
                        await websocket.send('PONG')
                    elif result_message == 'CONNECTED':
                        self._init_go_core = True
                        print(
                            f'{self.pytgcalls.OKGREEN} '
                            f'Started GO Core!'
                            f'{self.pytgcalls.ENDC}',
                        )
                    elif result_message != 'RECEIVED':
                        request_result = json.loads(result_message)
                        if 'path' in request_result:
                            asyncio.create_task(self.pytgcalls.http_request_manager(websocket, request_result))
                        else:
                            await websocket.send('RECEIVED')
        except ConnectionRefusedError:
            if limit > 0:
                await self._start_web_app(limit - 1)
        except ConnectionClosedError:
            pass