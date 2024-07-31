import sh

from . import Port, Protocol, daemon

MSG_ALLOWLIST_SUBNET_ADD_SUCCESS = "Subnet %s is allowlisted successfully."
MSG_ALLOWLIST_SUBNET_ADD_ERROR = "Subnet %s is already allowlisted."
MSG_ALLOWLIST_SUBNET_REMOVE_SUCCESS = "Subnet %s is removed from the allowlist successfully."
MSG_ALLOWLIST_SUBNET_REMOVE_ERROR = "Subnet %s is not allowlisted."

MSG_ALLOWLIST_PORT_ADD_SUCCESS = "Port %s (%s) is allowlisted successfully."
MSG_ALLOWLIST_PORT_ADD_ERROR = "Port %s (%s) is already allowlisted."
MSG_ALLOWLIST_PORT_REMOVE_SUCCESS = "Port %s (%s) is removed from the allowlist successfully."
MSG_ALLOWLIST_PORT_REMOVE_ERROR = "Port %s (%s) is not allowlisted."

MSG_ALLOWLIST_PORT_RANGE_ADD_SUCCESS = "Ports %s (%s) are allowlisted successfully."
MSG_ALLOWLIST_PORT_RANGE_REMOVE_SUCCESS = "Ports %s (%s) are removed from the allowlist successfully."
MSG_ALLOWLIST_PORT_RANGE_REMOVE_ERROR = "Ports %s (%s) are not allowlisted."

def add_ports_to_allowlist(ports_list: list[Port], allowlist_alias="allowlist"):
    for port in ports_list:
        if ":" in port.value:
            # Port range
            range_start, range_end = port.value.split(":")

            cmd = ["ports", range_start, range_end]
            if port.protocol != Protocol.ALL:
                cmd.extend(["protocol", str(port.protocol)])

            port_value = port.value.replace(":", " - ")
            expected_message = MSG_ALLOWLIST_PORT_RANGE_ADD_SUCCESS % (port_value, port.protocol)
        else:
            # Single port
            cmd = ["port", port.value]
            if port.protocol != Protocol.ALL:
                cmd.extend(["protocol", str(port.protocol)])

            port_value = port.value
            expected_message = MSG_ALLOWLIST_PORT_ADD_SUCCESS % (port_value, port.protocol)

        cmd_message = sh.nordvpn(allowlist_alias, "add", cmd)
        print(cmd_message)

        assert sh.nordvpn.settings().count(f" {port_value} ({str(port.protocol)})") == 1, \
            "Port(range) not found or found more than once in `nordvpn settings`"

        assert cmd_message is not None and expected_message in cmd_message, \
            f"Wrong allowlist message.\nExpected: {expected_message}\nGot: {cmd_message}"


def remove_ports_from_allowlist(ports_list: list[Port], allowlist_alias="allowlist"):
    for port in ports_list:
        if ":" in port.value:
            # Port range
            range_start, range_end = port.value.split(":")

            cmd = ["ports", range_start, range_end]
            if port.protocol != Protocol.ALL:
                cmd.extend(["protocol", str(port.protocol)])

            port_value = port.value.replace(":", " - ")
            expected_message = MSG_ALLOWLIST_PORT_RANGE_REMOVE_SUCCESS % (port_value, port.protocol)
        else:
            # Single port
            cmd = ["port", port.value]
            if port.protocol != Protocol.ALL:
                cmd.extend(["protocol", str(port.protocol)])

            port_value = port.value
            expected_message = MSG_ALLOWLIST_PORT_REMOVE_SUCCESS % (port_value, port.protocol)

        cmd_message = sh.nordvpn(allowlist_alias, "remove", cmd)
        print(cmd_message)

        assert sh.nordvpn.settings().count(f" {port_value} ({str(port.protocol)})") == 0, \
            "Port(range) found in `nordvpn settings`"

        assert cmd_message is not None and expected_message in cmd_message, \
            f"Wrong allowlist message.\nExpected: {expected_message}\nGot: {cmd_message}"


def add_subnet_to_allowlist(subnet_list: list[str], allowlist_alias="allowlist"):
    for subnet in subnet_list:
        cmd_message = sh.nordvpn(allowlist_alias, "add", "subnet", subnet)
        expected_message = MSG_ALLOWLIST_SUBNET_ADD_SUCCESS % subnet

        assert expected_message in cmd_message, \
            f"Wrong allowlist message.\nExpected: {expected_message}\nGot: {cmd_message}"

        assert sh.nordvpn.settings().count(subnet) == 1, \
            "Subnet not found or found more than once in `nordvpn settings`"

        # If subnet /32 whitelisted, only IP Address is visible in `ip route`
        if "/32" in subnet:
            subnet = subnet.replace("/32", "")  # noqa: PLW2901

        if daemon.is_connected():
            iprules = sh.ip.rule.show()
            assert subnet in iprules, f"Subnet {subnet} not found in `ip rule show`"


def remove_subnet_from_allowlist(subnet_list: list[str], allowlist_alias="allowlist"):
    for subnet in subnet_list:
        cmd_message = sh.nordvpn(allowlist_alias, "remove", "subnet", subnet)
        expected_message = MSG_ALLOWLIST_SUBNET_REMOVE_SUCCESS % subnet

        assert expected_message in cmd_message, \
            f"Wrong allowlist message.\nExpected: {expected_message}\nGot: {cmd_message}"

        assert sh.nordvpn.settings().count(subnet) == 0, \
            "Subnet found in `nordvpn settings`"

        # If subnet /32 whitelisted, only IP Address is visible in `ip route`
        if "/32" in subnet:
            subnet = subnet.replace("/32", "")  # noqa: PLW2901

        iprules = sh.ip.rule.show()
        assert subnet not in iprules, "Subnet found in `ip rule show`"
