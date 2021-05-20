import asyncio
import os
import socket
from time import time
from typing import Callable
from typing import Dict
from typing import List

from pyrogram import __version__
from pyrogram import Client
from pyrogram.raw.types import ChannelForbidden
from pyrogram.raw.types import GroupCall
from pyrogram.raw.types import GroupCallDiscarded
from pyrogram.raw.types import InputGroupCall
from pyrogram.raw.types import MessageActionInviteToGroupCall
from pyrogram.raw.types import UpdateChannel
from pyrogram.raw.types import UpdateGroupCall
from pyrogram.raw.types import UpdateNewChannelMessage

from .methods import Methods


class PyTgCalls(Methods):
    def __init__(
        self,
        app: Client,
        port: int = 24859,
        log_mode: int = 0,
        flood_wait_cache: int = 120,
    ):
        self._app = app
        self._app_core = None
        self._host = '127.0.0.1'
        self._port = port
        self._init_go_core = False
        self._on_event_update: Dict[str, list] = {
            'EVENT_UPDATE_HANDLER': [],
            'STREAM_END_HANDLER': [],
            'CUSTOM_API_HANDLER': [],
            'GROUP_CALL_HANDLER': [],
            'KICK_HANDLER': [],
            'CLOSED_HANDLER': [],
        }
        self._my_id = 0
        self.is_running = False
        self._calls: List[int] = []
        self._active_calls: Dict[int, str] = {}
        self._async_processes: Dict[str, Dict] = {}
        self._session_id = self._generate_session_id(20)
        self._log_mode = log_mode
        self._cache_user_peer: Dict[int, Dict] = {}
        self._cache_full_chat: Dict[int, Dict] = {}
        self._cache_local_peer = None
        self._flood_wait_cache = flood_wait_cache
        self._state_conn = 2
        self._gws = None
        super().__init__(self)

    @staticmethod
    def verbose_mode():
        return 1

    @property
    def ultra_verbose_mode(self):
        return 2

    @staticmethod
    def get_version(package_check):
        result_cmd = os.popen(f'{package_check} -v').read()
        result_cmd = result_cmd.replace('v', '')
        if len(result_cmd) == 0:
            return {
                'version_int': 0,
                'version': '0',
            }
        return {
            'version_int': int(result_cmd.split('.')[0]),
            'version': result_cmd,
        }

    def run(self, before_start_callable: Callable = None):
        if self._app is not None:
            if int(__version__.split('.')[1]) < 2 and \
                    int(__version__.split('.')[0]) == 1:
                raise Exception(
                    'Needed pyrogram 1.2.0+, '
                    'actually installed is '
                    f'{__version__}',
                )
            a_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            location = ('127.0.0.1', self._port)
            result_of_check = a_socket.connect_ex(location)
            is_already_in_use = result_of_check == 0
            a_socket.close()
            if not is_already_in_use:
                try:
                    # noinspection PyBroadException
                    @self._app.on_raw_update()
                    async def on_close(client, update, _, data2):
                        if isinstance(update, UpdateGroupCall):
                            if isinstance(update.call, GroupCallDiscarded):
                                chat_id = int(f'-100{update.chat_id}')
                                self._cache_full_chat[chat_id] = {
                                    'last_update': int(time()),
                                    'full_chat': None,
                                }
                            if isinstance(update.call, GroupCall):
                                input_group_call = InputGroupCall(
                                    access_hash=update.call.access_hash,
                                    id=update.call.id,
                                )
                                chat_id = int(f'-100{update.chat_id}')
                                self._cache_full_chat[chat_id] = {
                                    'last_update': int(time()),
                                    'full_chat': input_group_call,
                                }
                        if isinstance(update, UpdateChannel):
                            chat_id = int(f'-100{update.channel_id}')
                            if len(data2) > 0:
                                if isinstance(
                                        data2[update.channel_id],
                                        ChannelForbidden,
                                ):
                                    for event in self._on_event_update[
                                        'KICK_HANDLER'
                                    ]:
                                        await event['callable'](
                                            chat_id,
                                        )
                                    # noinspection PyBroadException
                                    try:
                                        self.leave_group_call(
                                            chat_id,
                                            'kicked_from_group',
                                        )
                                    except Exception:
                                        pass
                                    try:
                                        del self._cache_user_peer[chat_id]
                                    except Exception:
                                        pass
                        if isinstance(
                                update,
                                UpdateGroupCall,
                        ):
                            if isinstance(
                                    update.call,
                                    GroupCallDiscarded,
                            ):
                                chat_id = int(f'-100{update.chat_id}')
                                for event in self._on_event_update[
                                    'CLOSED_HANDLER'
                                ]:
                                    await event['callable'](
                                        chat_id,
                                    )
                                # noinspection PyBroadException
                                try:
                                    self.leave_group_call(
                                        chat_id,
                                        'closed_voice_chat',
                                    )
                                except Exception:
                                    pass
                                try:
                                    del self._cache_user_peer[chat_id]
                                except Exception:
                                    pass
                        if isinstance(
                                update,
                                UpdateNewChannelMessage,
                        ):
                            try:
                                if isinstance(
                                        update.message.action,
                                        MessageActionInviteToGroupCall,
                                ):
                                    for event in self._on_event_update[
                                        'GROUP_CALL_HANDLER'
                                    ]:
                                        await event['callable'](
                                            client, update.message,
                                        )
                            except Exception:
                                pass

                    self._app.start()
                    self._my_id = self._app.get_me()['id']  # noqa
                    self._cache_local_peer = self._app.resolve_peer(
                        self._my_id,
                    )
                    if before_start_callable is not None:
                        # noinspection PyBroadException
                        try:
                            result = before_start_callable(self._my_id)
                            if isinstance(result, bool):
                                if not result:
                                    return
                        except Exception:
                            pass
                    print(f'Starting on port: {self._port}')
                except KeyboardInterrupt:
                    pass
                try:
                    asyncio.get_event_loop().run_until_complete(self._prepare_run(
                        f'{__file__.replace("pytgcalls.py", "")}',
                        f'port={self._port} log_mode={self._log_mode}'
                    ))
                except KeyboardInterrupt:
                    print(
                        f'\n{self.pytgcalls.FAIL} '
                        f'GO Core Stopped, '
                        f'press Ctrl+C again to exit!'
                        f'{self.pytgcalls.ENDC}',
                    )
            else:
                raise OSError(
                    f'error while attempting to bind on address '
                    f'{location}: address already in use'
                )

        else:
            raise Exception('NEED_PYROGRAM_CLIENT')
        return self
    def _add_handler(self, type_event: str, func):
        self._on_event_update[type_event].append(func)
