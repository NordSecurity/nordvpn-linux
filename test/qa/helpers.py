import subprocess
import time


def search_journalctl_logs(search_pattern: str, since: str, service: str, timeout: int = 90, interval: int = 1) -> list:
    """
    Function to search for a pattern in the service logs since the given timestamp.

    :param  search_pattern:     String or regex pattern to search for
    :param  since:              Timestamp to search logs from
    :param  service:            Name of the service to check logs for
    :param  timeout:            Time to wait for logs to become available
    :param  interval:           Time to wait between checks

    :returns:                   List of found logs
    """
    time_mark = time.time()
    while time.time() < time_mark + timeout:
        try:
            cmd = ["journalctl"]

            if service:
                cmd += ["-u", service]
            if since:
                cmd += ["--since", since]

            result = subprocess.run(cmd, capture_output=True, text=True, check=True)

            if search_pattern:
                if hasattr(search_pattern, "search") and callable(search_pattern.search):
                    # Regex. Be sure to pass it like this -> re.compile(r"RC_LOAD_TIME_MIN=\d+")
                    matches = [line for line in result.stdout.splitlines() if search_pattern.search(line)]
                else:
                    # String. Whatever format
                    matches = [line for line in result.stdout.splitlines() if search_pattern in line]

                if matches:
                    print(f"Found match in {(time_mark + timeout) - time.time()} seconds.")
                    return matches

                print("Waiting for logs.")
                time.sleep(interval)
        except subprocess.CalledProcessError as e:
            print(f"Failed to get service logs for {service}: {e}")
            return []
    print(
        f"Couldn't find service logs for {service} in {timeout} seconds. Logs from that time are next: {result.stdout.splitlines()}"
    )
    return []
