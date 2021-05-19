from pyrogram.raw.functions.phone import EditGroupCallParticipant


class ChangeVolumeVoiceCall:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def _change_volume_voice_call(self, params: dict):
        result_json = {
            'result': 'ACCESS_DENIED',
        }
        if params['session_id'] == self.pytgcalls._session_id:
            # noinspection PyBroadException
            try:
                chat_call = await self.pytgcalls._load_chat_call(
                    params['chat_id'],
                )
                await self.pytgcalls._app.send(
                    EditGroupCallParticipant(
                        call=chat_call,
                        participant=self.pytgcalls._cache_user_peer[
                            int(params['chat_id'])
                        ],
                        muted=False,
                        volume=params['volume'] * 100,
                    ),
                )

                result_json = {
                    'result': 'OK',
                }
            except Exception:
                pass
        return result_json
