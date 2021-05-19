import asyncio
import json
from typing import Callable


class HTTPRequestManager:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def http_request_manager(self, websocket, data: dict):
        if data['path'] == '/request_join_call':
            await self._json_response(data['post'], data['session_id'],websocket, self.pytgcalls._join_voice_call)
        elif data['path'] == '/request_leave_call':
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._leave_voice_call)
        elif data['path'] == '/get_participants':
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._get_participants)
        elif data['path'] == '/ended_stream':
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._event_finish)
        elif data['path'] == '/update_request':
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._update_call_data)
        elif data['path'] == '/api_internal':
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._api_backend)
        elif data['path'] == '/request_change_volume':
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._change_volume_voice_call)
        elif data['path'] == '/async_request':
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._async_result)
        elif data['path'] == '/api' and len(self.pytgcalls._on_event_update['CUSTOM_API_HANDLER']) > 0:
            await self._json_response(data['post'], data['session_id'], websocket, self.pytgcalls._custom_api_update)
        else:
            await websocket.send(json.dumps({
                'status': {
                    'code': 404
                },
                'session_id': data['session_id']
            }))

    # noinspection PyBroadException
    @staticmethod
    async def _json_response(params: dict, session_id: str, websocket, c: Callable):
        try:
            result = await c(params)
            await websocket.send(json.dumps({
                'status': {
                    'code': 200
                },
                'result': result,
                'session_id': session_id
            }))
        except Exception:
            await websocket.send(json.dumps({
                'status': {
                    'code': 500
                },
                'session_id': session_id
            }))
