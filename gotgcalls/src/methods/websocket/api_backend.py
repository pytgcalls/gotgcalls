import json


class ApiBackend:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def _api_backend(self, params: dict):
        result_json = {
            'result': 'ACCESS_DENIED',
        }
        # noinspection PyBroadException
        try:
            if params['session_id'] == self.pytgcalls._session_id:
                await self.pytgcalls._gws.send(json.dumps(params))
                result_json = {
                    'result': 'ACCESS_GRANTED',
                }
        except Exception:
            pass
        return result_json
