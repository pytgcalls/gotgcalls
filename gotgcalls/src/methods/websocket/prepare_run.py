import os


class PrepareRun:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def _prepare_run(
            self,
            file_path: str,
            arguments: str
    ):
        os.system('color')
        self.pytgcalls._run_go(
            file_path,
            arguments
        )
        await self.pytgcalls._start_web_app()
