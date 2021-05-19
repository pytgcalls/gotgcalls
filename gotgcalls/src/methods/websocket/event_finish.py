class EventFinish:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def _event_finish(self, params: dict):
        chat_id = int(params['chat_id'])
        self.pytgcalls._remove_active_call(chat_id)

        for event in self.pytgcalls._on_event_update['STREAM_END_HANDLER']:
            self.pytgcalls.run_async(
                event['callable'],
                (chat_id,),
            )
        return {
            'result': 'OK',
        }
