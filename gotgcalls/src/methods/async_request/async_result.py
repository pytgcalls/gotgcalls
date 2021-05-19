class AsyncResult:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember,PyBroadException
    async def _async_result(self, params: dict):
        try:
            def_call = self.pytgcalls._async_processes[params['ID']]
            self.pytgcalls._async_processes[params['ID']] = {
                'RESULT': await def_call['CALLABLE'](*def_call['TUPLE']),
            }
        except Exception:
            pass
        del self.pytgcalls._async_processes[params['ID']]
        return {
            'result': 'OK',
        }
