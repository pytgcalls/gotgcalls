from typing import Callable


class OnMessage:
    def __init__(self, pytgcalls):
        self.pytgcalls = pytgcalls

    def _on_message(self) -> Callable:
        method = 'ON_SOCKET_MESSAGE'

        # noinspection PyProtectedMember
        def decorator(func: Callable) -> Callable:
            if self is not None:
                self.pytgcalls._add_handler(
                    method, {
                        'callable': func,
                    },
                )
            return func
        return decorator