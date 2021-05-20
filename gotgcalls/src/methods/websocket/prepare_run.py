import os
import sys


class PrepareRun:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    # noinspection PyProtectedMember
    async def _prepare_run(
            self,
            file_path: str,
            arguments: str
    ):
        is_windows = sys.platform.startswith('win')
        if is_windows:
            os.system('color')
        self.pytgcalls._spawn_process(
            self.pytgcalls._run_go,
            (
                file_path,
                arguments
            )
        )
        await self.pytgcalls._start_web_app()
