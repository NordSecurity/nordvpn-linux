import os
import time


class LogReader:
    """A utility class for reading logs, partially or fully, and waiting for specific messages in the log file."""

    def __init__(self, log_path: str):
        """
        Initializes the "LogReader" with the path to the log file.

        :param log_path: The path to the log file.
        """
        self.log_path = log_path

    def get_full_log(self) -> str:
        """
        Reads and returns the entire content of the log file.

        :return: The entire content of the log file as a string.

        :raises FileNotFoundError: If log_path is not set.
        :raises Exception: If an unexpected error occurs.
        """
        try:
            with open(self.log_path, encoding="utf-8", errors="ignore") as log_file:
                return log_file.read()
        except FileNotFoundError:
            print(f"Log file '{self.log_path}' not found.")
            raise
        except Exception as e:
            print(f"Error occurred while reading the log file: {e}")
            raise

    def get_partial_log(self, cursor: int) -> str:
        """
        Reads the log file starting from the given cursor position.

        :param cursor: The cursor position to start reading from.

        :return: The entire content of the log file as a string.

        :raises FileNotFoundError: If log_path is not set.
        :raises Exception: If an unexpected error occurs.
        """
        try:
            with open(self.log_path, "rb") as log_file:
                file_size = log_file.seek(0, 2)  # Move to the end to check file size

                # Validate the cursor (if greater than file size, cursor is invalid)
                if cursor > file_size:
                    print(
                        f"Cursor ({cursor}) exceeds file size ({file_size}). Returning empty content."
                    )
                    return ""

                # Move to the specified cursor position
                log_file.seek(cursor)

                content = log_file.read().decode("utf-8", errors="ignore")
                return content

        except FileNotFoundError:
            print(f"Log file '{self.log_path}' not found.")
            raise
        except Exception as e:
            print(f"Error occurred while reading part of the log file: {e}")
            raise

    def wait_for_messages(
        self,
        messages: list[str],
        cursor: int = 0,
        timeout: int = 30,
        interval: int = 1,
        return_not_found: bool = False,
    ) -> bool | tuple:
        """
        Wait for specific messages to appear in the log file within a defined timeout.

        This method continuously monitors a log file for the presence of specific messages, starting
        from the given cursor position. It checks for the first occurrence of each message, removes
        them from the tracking list, and waits for other messages to appear until either all are
        found or the timeout expires.

        :param messages: The messages to wait for.
        :param cursor: The cursor position to start reading from.
        :param timeout: The maximum number of seconds to wait for messages to appear.
        :param interval: The number of seconds to wait for messages to appear.
        :param return_not_found: If True, do not raise an exception when no messages are found.

        :return: bool. True if all messages are found within the timeout, False otherwise.

        :raises FileNotFoundError: If log file is not found.
        """
        start_time = time.time()
        messages_to_find = list(messages)

        print(f"Waiting for messages: {messages_to_find}")

        while time.time() - start_time < timeout:
            try:
                partial_log = self.get_partial_log(cursor)
                log_lines = partial_log.splitlines()

                for line in log_lines:
                    for message in messages_to_find:
                        if message in line:
                            print(f"Message found: '{message}' on line: {line.strip()}")
                            # Remove the message from the list after finding it
                            messages_to_find.remove(message)
                            break

                if not messages_to_find:
                    print("All specified messages have been found in the log.")
                    return True if not return_not_found else (True, messages_to_find)

            except FileNotFoundError:
                print(f"Log file '{self.log_path}' not found. Retrying...")

            time.sleep(interval)

        print(f"Messages not found within {timeout} seconds: {messages_to_find}")
        return False if not return_not_found else (False, messages_to_find)

    def get_cursor(self) -> int:
        """
        Gets the current EOF cursor for the log file. If the file does not exist, returns 0.

        :return: The current EOF cursor.

        :raises FileNotFoundError: If log file is not found.
        """
        cursor = 0
        try:
            with open(self.log_path, "rb") as log_file:
                log_file.seek(0, os.SEEK_END)  # Move to the end of the file
                cursor = log_file.tell()  # Get the current byte offset
        except FileNotFoundError:
            print(f"Log file '{self.log_path}' not found. Setting cursor to 0.")
        finally:
            print(f"cursor: {cursor}")
        return cursor
