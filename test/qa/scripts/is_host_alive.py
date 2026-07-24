import socket
import sys
import time

def is_host_alive(host, port, retries=3, delay=1):
    """Check if a host is reachable by attempting a TCP connection with retries."""
    for attempt in range(retries):
        try:
            with socket.create_connection((host, port), timeout=2):
                return True  # Port is open, host is alive
        except TimeoutError:
            pass  # No response, retry
        except ConnectionRefusedError:
            return True  # Port closed, but host is alive
        except OSError:
            pass  # No route, DNS failure, or unreachable network

        if attempt < retries - 1:
            time.sleep(delay)

    return False  # Exhausted retries, host is unreachable

if __name__ == "__main__":
    if len(sys.argv) == 5:
        print(is_host_alive(host=sys.argv[1], port=int(sys.argv[2]), retries=int(sys.argv[3]), delay=int(sys.argv[4])))
    else:
        print("Usage: python3 is_host_alive.py <host> <port> [retries] [delay]")
