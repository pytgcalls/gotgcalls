from pyrogram.raw.functions.phone import LeaveGroupCall


class LeaveVoiceCall:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def _leave_voice_call(self, params: dict):
        result = {
            'result': 'OK',
        }
        try:
            # noinspection PyBroadException
            chat_call = await self.pytgcalls._load_chat_call(
                int(params['chat_id']),
            )
            if chat_call is not None:
                # noinspection PyBroadException
                await self.pytgcalls._app.send(
                    LeaveGroupCall(
                        call=chat_call,
                        source=0,
                    ),
                )
        except Exception as e:
            result = {
                'result': str(e),
            }
        return result
