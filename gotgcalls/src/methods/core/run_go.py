import os


class RunGO:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    def _run_go(
            self,
            file_path: str = '',
            arguments: str = '',
    ):
        try:
            os.system(f'{file_path} {arguments}')
        except KeyboardInterrupt:
            self.is_running = False