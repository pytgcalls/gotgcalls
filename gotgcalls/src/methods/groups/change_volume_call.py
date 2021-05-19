import json

import requests

from ..core import SpawnProcess


class ChangeVolume(SpawnProcess):
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    def change_volume_call(self, chat_id: int, volume: int):
        volume = int(volume) if isinstance(volume, str) else volume
        if self.pytgcalls._init_go_core and \
                self.pytgcalls._app is not None and\
                chat_id in self.pytgcalls._cache_user_peer:
            volume = 200 if volume > 200 else (0 if volume < 0 else volume)
            try:
                self._spawn_process(
                    requests.post,
                    (
                        'http://'
                        f'{self.pytgcalls._host}:'
                        f'{self.pytgcalls._port}/'
                        'request_change_volume',
                        json.dumps({
                            'chat_id': chat_id,
                            'volume': volume,
                            'session_id': self.pytgcalls._session_id,
                        }),
                    ),
                )
            except Exception:
                raise Exception('Error internal: NOT_IN_GROUP')
        else:
            code_err = 'PYROGRAM_CLIENT_IS_NOT_RUNNING'
            if chat_id not in self.pytgcalls._cache_user_peer:
                code_err = 'GROUP_CALL_NOT_FOUND'
            if not self.pytgcalls._init_go_core:
                code_err = 'GO_CORE_NOT_RUNNING'
            raise Exception(f'Error internal: {code_err}')
