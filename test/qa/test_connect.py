import random
import socket
import time

import pytest
import sh

import lib
from lib import daemon, info, logging, network, server

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


CONNECT_ALIAS = [
    "connect",
    "c"
]


def get_alias() -> str:
    """
    This function randomly picks an alias from the predefined list 'CONNECT_ALIAS' and returns it.

    Returns:
        str: A randomly selected alias from CONNECT_ALIAS.
    """
    return random.choice(CONNECT_ALIAS)


def connect_base_test(connection_settings, group=(), name="", hostname=""):
    print(connection_settings)
    output = sh.nordvpn(get_alias(), group, _tty_out=False)
    print(output)

    assert lib.is_connect_successful(output, name, hostname)
    assert network.is_available()


def disconnect_base_test():
    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()
    assert "nordlynx" not in sh.ip.a() and "nordtun" not in sh.ip.a() and "qtun" not in sh.ip.a()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_quick_connect(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-773
        [openvpn-tcp-on] - LVPN-829
        [openvpn-udp-off] - LVPN-557
        [openvpn-tcp-off] - LVPN-559
        [nordlynx--] - LVPN-711
        [nordwhisper--] - LVPN-6717
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated))
    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_server_absent(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-8667
        [openvpn-tcp-on] - LVPN-8666
        [openvpn-udp-off] - LVPN-8669
        [openvpn-tcp-off] - LVPN-8668
        [nordlynx--] - LVPN-8671
        [nordwhisper--] - LVPN-8670
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "moon")

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)
    assert network.is_disconnected()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_server_random_by_name(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-637
        [openvpn-tcp-on] - LVPN-755
        [openvpn-udp-off] - LVPN-5816
        [openvpn-tcp-off] - LVPN-5800
        [nordlynx--] - LVPN-5817
        [nordwhisper--] - LVPN-6721
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_group_random_server_by_name_additional(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-off-Double_VPN] - LVPN-8661
        [openvpn-udp-off-Onion_Over_VPN] - LVPN-8846
        [openvpn-udp-off-Standard_VPN_Servers] - LVPN-8845
        [openvpn-udp-off-P2P] - LVPN-8844

        [openvpn-tcp-off-Double_VPN] - LVPN-8847
        [openvpn-tcp-off-Onion_Over_VPN] - LVPN-8848
        [openvpn-tcp-off-Standard_VPN_Servers] - LVPN-8849
        [openvpn-tcp-off-P2P] - LVPN-8662

        [nordlynx---Double_VPN] - LVPN-8850
        [nordlynx---Onion_Over_VPN] - LVPN-8851
        [nordlynx---Standard_VPN_Servers] - LVPN-8852
        [nordlynx---P2P] - LVPN-8649
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)

    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS_NORDWHISPER)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
def test_nordwhisper_connect_to_group_random_server_by_name_additional(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [nordwhisper---Standard_VPN_Servers] - LVPN-8855
        [nordwhisper---P2P] - LVPN-8663
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)

    disconnect_base_test()


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_group_random_server_by_name_standard(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-on-Africa_The_Middle_East_And_India] - LVPN-8861
        [openvpn-udp-on-Asia_Pacific] - LVPN-8644
        [openvpn-udp-on-The_Americas] - LVPN-8866
        [openvpn-udp-on-Europe] - LVPN-8870

        [openvpn-tcp-on-Africa_The_Middle_East_And_India] - LVPN-8862
        [openvpn-tcp-on-Asia_Pacific] - LVPN-8642
        [openvpn-tcp-on-The_Americas] - LVPN-8867
        [openvpn-tcp-on-Europe] - LVPN-8871

        [openvpn-udp-off-Africa_The_Middle_East_And_India] - LVPN-8863
        [openvpn-udp-off-Asia_Pacific] - LVPN-8646
        [openvpn-udp-off-The_Americas] - LVPN-8868
        [openvpn-udp-off-Europe] - LVPN-8872

        [openvpn-tcp-off-Africa_The_Middle_East_And_India] - LVPN-8864
        [openvpn-tcp-off-Asia_Pacific] - LVPN-8645
        [openvpn-tcp-off-The_Americas] - LVPN-8869
        [openvpn-tcp-off-Europe] - LVPN-8873

        [nordlynx---Africa_The_Middle_East_And_India] - LVPN-8860
        [nordlynx---Asia_Pacific] - LVPN-8648
        [nordlynx---The_Americas] - LVPN-8865
        [nordlynx---Europe] - LVPN-8874

        [nordwhisper---Africa_The_Middle_East_And_India] - LVPN-8859
        [nordwhisper---Asia_Pacific] - LVPN-8647
        [nordwhisper---The_Americas] - LVPN-8858
        [nordwhisper---Europe] - LVPN-8857
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)

    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_to_group_random_server_by_name_obfuscated(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-8665
        [openvpn-tcp-on] - LVPN-8642
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(group_name=group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)
    disconnect_base_test()


@pytest.mark.skip("flaky test, LVPN-6277")
# the tun interface is recreated only for OpenVPN
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_network_restart_recreates_tun_interface(tech, proto, obfuscated):
    """Manual TC is unavailable because reconnection timing and interface changes can’t be reliably checked without automation."""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated))

    links = socket.if_nameindex()
    logging.log(links)
    default_gateway = network.stop()
    network.start(default_gateway)
    daemon.wait_for_reconnect(links)
    assert network.is_connected()
    logging.log(info.collect())

    disconnect_base_test()


# for Nordlynx normally the tunnel is not recreated
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1)
def test_connect_network_restart_nordlynx(tech, proto, obfuscated):
    """Manual TC is unavailable because reconnection timing and interface changes can’t be reliably checked without automation."""

    if daemon.is_init_systemd():
        pytest.skip("LVPN-5733")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated))

    links = socket.if_nameindex()
    logging.log(links)
    default_gateway = network.stop()
    network.start(default_gateway)

    # wait for internet
    network.is_available(10)

    assert network.is_connected()
    assert links == socket.if_nameindex()

    logging.log(info.collect())

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_quick_connect_double_disconnect(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-1059
        [openvpn-tcp-on] - LVPN-1060
        [openvpn-udp-off] - LVPN-1057
        [openvpn-tcp-off] - LVPN-1058
        [nordlynx--] - LVPN-1056
        [nordwhisper--] - LVPN-8066
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    for _ in range(2):
        connect_base_test((tech, proto, obfuscated))
        disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_network_gone(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-5814
        [openvpn-tcp-on] - LVPN-5805
        [openvpn-udp-off] - LVPN-5795
        [openvpn-tcp-off] - LVPN-5796
        [nordlynx--] - LVPN-911
        [nordwhisper--] - LVPN-8065
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    default_gateway = network.stop()
    with lib.Defer(lambda: network.start(default_gateway)):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(get_alias())
        print(ex.value)


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_group_standard(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-on-Africa_The_Middle_East_And_India] - LVPN-750
        [openvpn-udp-on-Asia_Pacific] - LVPN-742
        [openvpn-udp-on-The_Americas] - LVPN-757
        [openvpn-udp-on-Europe] - LVPN-770

        [openvpn-tcp-on-Africa_The_Middle_East_And_India] - LVPN-748
        [openvpn-tcp-on-Asia_Pacific] - LVPN-744
        [openvpn-tcp-on-The_Americas] - LVPN-756
        [openvpn-tcp-on-Europe] - LVPN-764

        [openvpn-udp-off-Africa_The_Middle_East_And_India] - LVPN-811
        [openvpn-udp-off-Asia_Pacific] - LVPN-833
        [openvpn-udp-off-The_Americas] - LVPN-798
        [openvpn-udp-off-Europe] - LVPN-831

        [openvpn-tcp-off-Africa_The_Middle_East_And_India] - LVPN-809
        [openvpn-tcp-off-Asia_Pacific] - LVPN-836
        [openvpn-tcp-off-The_Americas] - LVPN-799
        [openvpn-tcp-off-Europe] - LVPN-823

        [nordlynx---Africa_The_Middle_East_And_India] - LVPN-518
        [nordlynx---Asia_Pacific] - LVPN-495
        [nordlynx---The_Americas] - LVPN-433
        [nordlynx---Europe] - LVPN-470

        [nordwhisper---Africa_The_Middle_East_And_India] - LVPN-8073
        [nordwhisper---Asia_Pacific] - LVPN-8072
        [nordwhisper---The_Americas] - LVPN-8068
        [nordwhisper---Europe] - LVPN-8071
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_group_additional(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-off-Double_VPN] - LVPN-837
        [openvpn-udp-off-Onion_Over_VPN] - LVPN-824
        [openvpn-udp-off-Standard_VPN_Servers] - LVPN-775
        [openvpn-udp-off-P2P] - LVPN-780

        [openvpn-tcp-off-Double_VPN] - LVPN-838
        [openvpn-tcp-off-Onion_Over_VPN] - LVPN-819
        [openvpn-tcp-off-Standard_VPN_Servers] - LVPN-806
        [openvpn-tcp-off-P2P] - LVPN-781

        [nordlynx---Double_VPN] - LVPN-469
        [nordlynx---Onion_Over_VPN] - LVPN-475
        [nordlynx---Standard_VPN_Servers] - LVPN-432
        [nordlynx---P2P] - LVPN-455
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS_NORDWHISPER)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
def test_nordwhisper_connect_to_group_additional(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [nordwhisper---Standard_VPN_Servers] - LVPN-8069
        [nordwhisper---P2P] - LVPN-8070
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.DEDICATED_IP_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_group_ovpn(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-off-Dedicated_IP] - LVPN-751
        [openvpn-tcp-off-Dedicated_IP] - LVPN-668
        [nordlynx---Dedicated_IP] - LVPN-5818
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_to_group_obfuscated(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-on-Obfuscated_Servers] - LVPN-762
        [openvpn-tcp-on-Obfuscated_Servers] - LVPN-768
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_flag_group_standard(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-on-Africa_The_Middle_East_And_India] - LVPN-8621
        [openvpn-udp-on-Asia_Pacific] - LVPN-8911
        [openvpn-udp-on-The_Americas] - LVPN-8903
        [openvpn-udp-on-Europe] - LVPN-8897

        [openvpn-tcp-on-Africa_The_Middle_East_And_India] - LVPN-8622
        [openvpn-tcp-on-Asia_Pacific] - LVPN-8896
        [openvpn-tcp-on-The_Americas] - LVPN-8895
        [openvpn-tcp-on-Europe] - LVPN-8894

        [openvpn-udp-off-Africa_The_Middle_East_And_India] - LVPN-8619
        [openvpn-udp-off-Asia_Pacific] - LVPN-8909
        [openvpn-udp-off-The_Americas] - LVPN-8905
        [openvpn-udp-off-Europe] - LVPN-8899

        [openvpn-tcp-off-Africa_The_Middle_East_And_India] - LVPN-8620
        [openvpn-tcp-off-Asia_Pacific] - LVPN-8910
        [openvpn-tcp-off-The_Americas] - LVPN-8904
        [openvpn-tcp-off-Europe] - LVPN-8898

        [nordlynx---Africa_The_Middle_East_And_India] - LVPN-8617
        [nordlynx---Asia_Pacific] - LVPN-8908
        [nordlynx---The_Americas] - LVPN-8902
        [nordlynx---Europe] - LVPN-8900

        [nordwhisper---Africa_The_Middle_East_And_India] - LVPN-8618
        [nordwhisper---Asia_Pacific] - LVPN-8907
        [nordwhisper---The_Americas] - LVPN-8906
        [nordwhisper---Europe] - LVPN-8901
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_flag_group_additional(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-off-Double_VPN] - LVPN-8616
        [openvpn-udp-off-Onion_Over_VPN] - LVPN-8915
        [openvpn-udp-off-Standard_VPN_Servers] - LVPN-8918
        [openvpn-udp-off-P2P] - LVPN-8922

        [openvpn-tcp-off-Double_VPN] - LVPN-8615
        [openvpn-tcp-off-Onion_Over_VPN] - LVPN-8914
        [openvpn-tcp-off-Standard_VPN_Servers] - LVPN-8913
        [openvpn-tcp-off-P2P] - LVPN-8912

        [nordlynx---Double_VPN] - LVPN-8613
        [nordlynx---Onion_Over_VPN] - LVPN-8916
        [nordlynx---Standard_VPN_Servers] - LVPN-8919
        [nordlynx---P2P] - LVPN-8921
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS_NORDWHISPER)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
def test_nordwhisper_connect_to_flag_group_additional(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [nordwhisper---Standard_VPN_Servers] - LVPN-8614
        [nordwhisper---P2P] - LVPN-8917
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.DEDICATED_IP_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_flag_group_ovpn(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-off-Dedicated_IP] - LVPN-8625
        [openvpn-tcp-off-Dedicated_IP] - LVPN-8623
        [nordlynx---Dedicated_IP] - LVPN-8624
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_to_flag_group_obfuscated(tech, proto, obfuscated, group):
    """
    Manual TCs:

        [openvpn-udp-on-Obfuscated_Servers] - LVPN-8630
        [openvpn-tcp-on-Obfuscated_Servers] - LVPN-8629
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_group_invalid(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-8632
        [openvpn-tcp-on] - LVPN-8631
        [openvpn-udp-off] - LVPN-8634
        [openvpn-tcp-off] - LVPN-8633
        [nordlynx--] - LVPN-8636
        [nordwhisper--] - LVPN-8635
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", "nonexistent_group")

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("country", lib.COUNTRIES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_country(tech, proto, obfuscated, country):
    """
    Manual TCs:

        [openvpn-udp-on-Germany/Netherlands/United_States/France] - LVPN-5806
        [openvpn-tcp-on-Germany/Netherlands/United_States/France] - LVPN-5797
        [openvpn-udp-off-Germany/Netherlands/United_States/France] - LVPN-487
        [openvpn-tcp-off-Germany/Netherlands/United_States/France] - LVPN-489
        [nordlynx---Germany/Netherlands/United_States/France] - LVPN-682
        [nordwhisper---Germany/Netherlands/United_States/France] - LVPN-6718
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), country)
    disconnect_base_test()


@pytest.mark.parametrize("country_code", lib.COUNTRY_CODES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_country_code(tech, proto, obfuscated, country_code):
    """
    Manual TCs:

        [openvpn-udp-on-de/nl/us/fr] - LVPN-5807
        [openvpn-tcp-on-de/nl/us/fr] - LVPN-5798
        [openvpn-udp-off-de/nl/us/fr] - LVPN-681
        [openvpn-tcp-off-de/nl/us/fr] - LVPN-843
        [nordlynx---de/nl/us/fr] - LVPN-678
        [nordwhisper---de/nl/us/fr] - LVPN-6719
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), country_code)
    disconnect_base_test()


@pytest.mark.parametrize("city", lib.CITIES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_city(tech, proto, obfuscated, city):
    """
    Manual TCs:

        [openvpn-udp-on-Frankfurt/Amsterdam/New_York/Paris] - LVPN-5808
        [openvpn-tcp-on-Frankfurt/Amsterdam/New_York/Paris] - LVPN-5799
        [openvpn-udp-off-Frankfurt/Amsterdam/New_York/Paris] - LVPN-844
        [openvpn-tcp-off-Frankfurt/Amsterdam/New_York/Paris] - LVPN-815
        [nordlynx---Frankfurt/Amsterdam/New_York/Paris] - LVPN-685
        [nordwhisper---Frankfurt/Amsterdam/New_York/Paris] - LVPN-6720
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), city)
    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_unavailable_groups(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-5810
        [openvpn-tcp-on] - LVPN-5801
        [openvpn-udp-off] - LVPN-8516
        [openvpn-tcp-off] - LVPN-8517
        [nordlynx--] - LVPN-8511
        [nordwhisper--] - LVPN-8513
    """

    # TODO: LVPN-257
    time.sleep(3)

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(get_alias(), group)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_connect_to_unavailable_servers(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-635
        [openvpn-tcp-on] - LVPN-636
        [openvpn-udp-off] - LVPN-424
        [openvpn-tcp-off] - LVPN-422
        [nordlynx--] - LVPN-774
    """

    # TODO: LVPN-257
    time.sleep(3)

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        server_info = server.get_hostname_by(group_name=group)
        name = server_info.hostname.split(".")[0]

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(get_alias(), name)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_status_connected(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-on] - LVPN-686
        [openvpn-tcp-on] - LVPN-687
        [openvpn-udp-off] - LVPN-675
        [openvpn-tcp-off] - LVPN-676
        [nordlynx--] - LVPN-841
        [nordwhisper--] - LVPN-8541
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert network.is_disconnected()
    assert "Disconnected" in sh.nordvpn.status()

    server_info = server.get_hostname_by(technology=tech, protocol=proto, obfuscated=obfuscated)
    sh.nordvpn(get_alias(), server_info.hostname.split(".")[0])

    connect_time = time.monotonic()

    network.generate_traffic(retry=5)

    status_info = daemon.get_status_data()
    status_time = time.monotonic()

    print("status_info: " + str(status_info))
    print("actual_status: " + str(sh.nordvpn.status()))

    assert "Connected" in status_info["status"]

    assert server_info.hostname in status_info["hostname"]
    assert server_info.name in status_info["server"]

    assert socket.gethostbyname(server_info.hostname) in status_info["ip"]

    assert server_info.country in status_info["country"]
    assert server_info.city in status_info["city"]

    assert tech.upper() in status_info["current technology"]

    if tech == "openvpn":
        assert proto.upper() in status_info["current protocol"]
    elif tech == "nordwhisper":
        assert "Webtunnel" in status_info["current protocol"]
    else:
        assert "UDP" in status_info["current protocol"]

    transfer_received = float(status_info["transfer"].split(" ")[0])
    transfer_sent = float(status_info["transfer"].split(" ")[3])

    assert transfer_received >= 0
    assert transfer_sent > 0

    time_connected = int(status_info["uptime"].split(" ")[0])
    time_passed = status_time - connect_time
    if "minute" in status_info["uptime"]:
        time_connected_seconds = int(status_info["uptime"].split(" ")[2])
        assert time_passed - 1 <= time_connected * 60 + time_connected_seconds <= time_passed + 1
    else:
        assert time_passed - 1 <= time_connected <= time_passed + 1

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_connect_to_virtual_server(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-off] - LVPN-5317
        [openvpn-tcp-off] - LVPN-5316
        [nordlynx--] - LVPN-5262
        [nordwhisper--] - LVPN-8531
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set("virtual-location", "on")
    virtual_countries = lib.get_virtual_countries()

    assert len(virtual_countries) > 0
    country = random.choice(virtual_countries)

    connect_base_test((tech, proto, obfuscated), country)
    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1)
def test_connect_to_post_quantum_server(tech, proto, obfuscated):
    """Manual TC: LVPN-5794"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.pq("on")

    connect_base_test((tech, proto, obfuscated))

    assert "preshared key" in sh.sudo.wg.show()

    disconnect_base_test()


def test_check_routing_table_for_lan():
    """Manual TC: LVPN-8728"""

    # check that the routing table is correctly configured when LAN is enabled and that the tunnel IP is correct
    lib.set_technology_and_protocol("nordlynx", "", "")

    default_route = network.RouteInfo.default_route_info()
    connect_base_test(("nordlynx", "", ""))

    # check the tunnel IP
    nordlynx_route = network.RouteInfo(sh.ip.route.show.dev("nordlynx").stdout.decode())
    assert nordlynx_route.destination == "10.5.0.0/16"
    assert nordlynx_route.src == "10.5.0.2"

    lan_ips = [
        "10.0.0.1",
        "172.16.0.1",
        "192.168.0.1",
        "169.254.0.1",
    ]

    private_vpn_ip = "10.5.0.1"

    # check LAN IP is not routed thru main
    assert all(not default_route.routes_ip(ip) for ip in lan_ips)
    assert not default_route.routes_ip(private_vpn_ip)
    assert nordlynx_route.routes_ip(private_vpn_ip)

    # enable LAN discovery
    sh.nordvpn.set("lan-discovery", "on")

    # check LAN IP is routed thru main
    assert all(default_route.routes_ip(ip) for ip in lan_ips)

    # IP from VPN private range is routed thru nordlynx interface
    assert not default_route.routes_ip(private_vpn_ip)
    assert nordlynx_route.routes_ip(private_vpn_ip)

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_dedicated_ip(tech, proto, obfuscated):
    """
    Manual TCs:

        [openvpn-udp-off] - LVPN-652
        [openvpn-tcp-off] - LVPN-651
        [nordlynx--] - LVPN-5819
    """

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_dedicated_ip()

    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)
    disconnect_base_test()

    assert network.is_disconnected()
    assert "nordlynx" not in sh.ip.a() and "nordtun" not in sh.ip.a()
