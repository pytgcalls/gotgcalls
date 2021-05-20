import subprocess
import sys


class RunGO:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    def _run_go(
            self,
            file_path: str = '',
            arguments: str = '',
    ):
        try:
            is_windows = sys.platform.startswith('win')
            if is_windows:
                file_executable = 'core.exe'
            else:
                file_executable = './core'
            subprocess.Popen(f'{file_path}{file_executable} {arguments}', creationflags=8, close_fds=True)
        except KeyboardInterrupt:
            self.pytgcalls.is_running = False