import sh

import time

from . import logging

class ShWithoutTTY:
    def __init__(self, base=None, level=None):
        self.base = base or sh
        self.level = level

    def __getattr__(self, name):
        # Create a new wrapper for the next level
        return ShWithoutTTY(base=getattr(self.base, name), level=name)

    def __call__(self, *args, **kwargs):
        # When the command is finally called, add _tty_out=False
        kwargs["_tty_out"] = False
        start = time.time()
        result = self.base(*args, **kwargs)
        end = time.time()
        exec_diff = end - start
        logging.log(f"[TEST_SUITE] SH '{self.level}' took {exec_diff:.3f} seconds")
        return result

sh_no_tty = ShWithoutTTY()
