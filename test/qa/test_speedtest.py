import json
import os
from datetime import datetime
from statistics import median

import pytest
import sh

import lib
from lib import network

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")

MIN_DOWNLOAD_MBPS = 70
MIN_UPLOAD_MBPS = 50
MAX_PING_MS = 50
MIN_SPEED_RATIO = 0.25

DOWNLOAD_BYTES = 5000000
UPLOAD_BYTES = 5000000
CLOUDFLARE_DOWN_URL = f"https://speed.cloudflare.com/__down?bytes={DOWNLOAD_BYTES}"
CLOUDFLARE_UP_URL = "https://speed.cloudflare.com/__up"
PING_HOST = "1.1.1.1"
PING_COUNT = 5
UPLOAD_FILE = "/tmp/speedtest_upload.bin"

SPEED_TEST_RUNS = 3

ARTIFACTS_DIR = os.path.join(os.environ.get("WORKDIR", "."), "dist", "test_artifacts")


@pytest.fixture(scope="session")
def speedtest_results_path():
    """Return path to the shared JSON file."""
    os.makedirs(ARTIFACTS_DIR, exist_ok=True)
    path = os.path.join(ARTIFACTS_DIR, "speedtest_results.json")
    if os.path.exists(path):
        os.remove(path)
    return path


def _write_results(path: str, results: dict):
    """Append measurement results to JSON file."""
    existing = []
    if os.path.exists(path):
        with open(path) as f:
            existing = json.load(f)

    existing.append(results)

    with open(path, "w") as f:
        json.dump(existing, f, indent=2, default=str)


def _measure_ping() -> float:
    """Measure ICMP ping to Cloudflare DNS, return average latency in ms."""
    output = sh.ping("-c", str(PING_COUNT), "-q", PING_HOST, _timeout=15)
    # parse "rtt min/avg/max/mdev = 1.234/5.678/9.012/1.234 ms"
    for line in str(output).splitlines():
        if "avg" in line:
            stats = line.split("=")[1].strip().split("/")
            return float(stats[1])
    msg = f"Could not parse ping output: {output}"
    raise ValueError(msg)


def _measure_download() -> float:
    """Download test file from Cloudflare, return median speed in Mbps."""
    results = []
    for _ in range(SPEED_TEST_RUNS):
        output = sh.curl(
            "-o", "/dev/null",
            "-s", "-w", "%{speed_download}",
            CLOUDFLARE_DOWN_URL,
            _timeout=60,
            _ok_code=[0, 56],
        )
        results.append(float(str(output).strip()) * 8 / 1000000)

    return median(results)


def _measure_upload() -> float:
    """Generate temp file and upload to Cloudflare, return median speed in Mbps."""
    sh.dd(
        "if=/dev/zero",
        f"of={UPLOAD_FILE}",
        "bs=1M",
        f"count={UPLOAD_BYTES // 1000000}",
        _err="/dev/null",
    )

    results = []
    for _ in range(SPEED_TEST_RUNS):
        output = sh.curl(
            "-X", "POST",
            "-o", "/dev/null",
            "-s", "-w", "%{speed_upload}",
            "--data-binary", f"@{UPLOAD_FILE}",
            CLOUDFLARE_UP_URL,
            _timeout=60,
            _ok_code=[0, 56],
        )
        results.append(float(str(output).strip()) * 8 / 1000000)

    if os.path.exists(UPLOAD_FILE):
        os.remove(UPLOAD_FILE)

    return median(results)


def _run_speedtest() -> dict:
    """Measure ping, download, and upload speed."""
    ping_ms = _measure_ping()
    download_mbps = _measure_download()
    upload_mbps = _measure_upload()
    return {
        "download_mbps": round(download_mbps, 2),
        "upload_mbps": round(upload_mbps, 2),
        "ping_ms": round(ping_ms, 2),
    }


def test_speed_thresholds(speedtest_results_path):
    """
    Test to verify internet connection speed meets minimum thresholds while connected via NordLynx.

    Measure download speed, upload speed, and ping while connected to VPN using NordLynx protocol.

    Test steps:
      1. Set technology to NordLynx.
      2. Connect to VPN.
      3. Run speedtest measurement with VPN.
      4. Disconnect from VPN.
      5. Verify download/upload speed and ping metrics.

    :raises AssertionError: If download speed is below MIN_DOWNLOAD_MBPS.
    :raises AssertionError: If upload speed is below MIN_UPLOAD_MBPS.
    :raises AssertionError: If ping latency exceeds MAX_PING_MS.
    """
    tech = "nordlynx"
    lib.set_technology_and_protocol(tech, "", "")

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert network.is_available(), "Network should be available when connected"

        vpn_data = _run_speedtest()

    _write_results(speedtest_results_path, {
        "test_name": "test_speed_nordlynx",
        "timestamp": datetime.now().isoformat(),
        "download_mbps": vpn_data["download_mbps"],
        "upload_mbps": vpn_data["upload_mbps"],
        "ping_ms": vpn_data["ping_ms"],
    })

    assert vpn_data["download_mbps"] >= MIN_DOWNLOAD_MBPS, (
        f"NordLynx download {vpn_data['download_mbps']} Mbps is below minimum {MIN_DOWNLOAD_MBPS} Mbps"
    )
    assert vpn_data["upload_mbps"] >= MIN_UPLOAD_MBPS, (
        f"NordLynx upload {vpn_data['upload_mbps']} Mbps is below minimum {MIN_UPLOAD_MBPS} Mbps"
    )
    assert vpn_data["ping_ms"] <= MAX_PING_MS, (
        f"NordLynx ping {vpn_data['ping_ms']} ms exceeds maximum {MAX_PING_MS} ms"
    )

def test_speed_degradation(speedtest_results_path):
    """
    Test to verify no significant speed degradation occurs after connecting to VPN via NordLynx.

    Measure baseline internet speed without VPN, then measure speed while connected via NordLynx protocol,
    and assert that VPN speed does not drop below required level 50% of the baseline.

    Test steps:
      1. Run speedtest measurement without VPN.
      2. Set technology to NordLynx.
      3. Connect to VPN.
      4. Run speedtest measurement with VPN.
      5. Disconnect from VPN.
      6. Verify download and upload speed ratios (VPN / baseline).

    :raises AssertionError: If VPN download speed drops below MIN_SPEED_RATIO of baseline.
    :raises AssertionError: If VPN upload speed drops below MIN_SPEED_RATIO of baseline.
    """
    baseline_data = _run_speedtest()

    tech = "nordlynx"
    lib.set_technology_and_protocol(tech, "", "")

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert network.is_available(), "Network should be available when connected"

        vpn_data = _run_speedtest()

    dl_ratio = round(vpn_data["download_mbps"] / baseline_data["download_mbps"], 2)
    ul_ratio = round(vpn_data["upload_mbps"] / baseline_data["upload_mbps"], 2)

    _write_results(speedtest_results_path, {
        "test_name": "test_speed_degradation",
        "timestamp": datetime.now().isoformat(),
        "baseline": {
            "download_mbps": baseline_data["download_mbps"],
            "upload_mbps": baseline_data["upload_mbps"],
            "ping_ms": baseline_data["ping_ms"],
        },
        tech: {
            "download_mbps": vpn_data["download_mbps"],
            "upload_mbps": vpn_data["upload_mbps"],
            "ping_ms": vpn_data["ping_ms"],
        },
        "download_ratio": dl_ratio,
        "upload_ratio": ul_ratio,
    })

    assert vpn_data["download_mbps"] >= baseline_data["download_mbps"] * MIN_SPEED_RATIO, (
        f"Download regression: {vpn_data['download_mbps']} Mbps is {dl_ratio * 100:.0f}% of "
        f"baseline ({baseline_data['download_mbps']} Mbps), minimum allowed is {MIN_SPEED_RATIO * 100:.0f}%"
    )
    assert vpn_data["upload_mbps"] >= baseline_data["upload_mbps"] * MIN_SPEED_RATIO, (
        f"Upload regression: {vpn_data['upload_mbps']} Mbps is {ul_ratio * 100:.0f}% of "
        f"baseline ({baseline_data['upload_mbps']} Mbps), minimum allowed is {MIN_SPEED_RATIO * 100:.0f}%"
    )
