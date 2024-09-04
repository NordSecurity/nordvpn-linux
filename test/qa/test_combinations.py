import pytest
import sh

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
)
from test_connect import disconnect_base_test, get_alias
from test_connect6 import connect_base_test


def setup_function(function):  # noqa: ARG001
    daemon.start()
    login.login_as("default")
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()

    sh.nordvpn.logout("--persist-token")
    sh.nordvpn.set.defaults()
    daemon.stop()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.TECHNOLOGIES)
def test_reconnect_matrix(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)
    connect_base_test(ipv6 = False)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)
    connect_base_test(ipv6 = False)

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize(("country", "city"), list(zip(lib.COUNTRIES, lib.CITIES, strict=False)))
def test_connect_country_and_city(tech, proto, obfuscated, country, city):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(country, ipv6 = False)
    connect_base_test(city, ipv6 = False)
    connect_base_test(f"{country} {city}", ipv6 = False)

    disconnect_base_test()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_status_change_technology_and_protocol(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    sh.nordvpn(get_alias())
    assert source_tech.upper() in sh.nordvpn.status()

    if source_tech == "openvpn":
        assert source_proto.upper() in sh.nordvpn.status()
    else:
        assert "UDP" in sh.nordvpn.status()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)
    assert source_tech.upper() in sh.nordvpn.status()

    if source_tech == "openvpn":
        assert source_tech.upper() in sh.nordvpn.status()
    else:
        assert "UDP" in sh.nordvpn.status()

    disconnect_base_test()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_status_change_technology_and_protocol_reconnect(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)
    sh.nordvpn(get_alias())
    disconnect_base_test()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)
    sh.nordvpn(get_alias())

    assert target_tech.upper() in sh.nordvpn.status()

    if target_tech == "openvpn":
        assert target_proto.upper() in sh.nordvpn.status()
    else:
        assert "UDP" in sh.nordvpn.status()

    disconnect_base_test()


@pytest.mark.parametrize("source_group", lib.STANDARD_GROUPS[-2:])
@pytest.mark.parametrize("target_group", lib.STANDARD_GROUPS[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_reconnect_to_standard_group(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
    source_group,
    target_group,
):

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test(source_group, ipv6 = False)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test(target_group, ipv6 = False)

    disconnect_base_test()


@pytest.mark.parametrize("source_group", lib.ADDITIONAL_GROUPS[-2:])
@pytest.mark.parametrize("target_group", lib.ADDITIONAL_GROUPS[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_reconnect_to_additional_group(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
    source_group,
    target_group,
):

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test(source_group, ipv6 = False)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test(target_group, ipv6 = False)

    disconnect_base_test()


@pytest.mark.parametrize("source_country", lib.COUNTRIES[-2:])
@pytest.mark.parametrize("target_country", lib.COUNTRIES[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_reconnect_to_server_by_country_name(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
    source_country,
    target_country,
):

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test(source_country, ipv6 = False)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test(target_country, ipv6 = False)

    disconnect_base_test()
