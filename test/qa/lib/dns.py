from lib import logging
import sh

# returns True when NordVPN DNS servers are used
def _is_set_to_nord():
    return "103.86.96.100" and "103.86.99.100" in sh.cat("/etc/resolv.conf")


# returns True when Threat Protection Lite DNS servers are used
def _is_set_to_threat_protection_lite():
    return "103.86.96.96" and "103.86.99.99" in sh.cat("/etc/resolv.conf")


# returns True when NordVPN DNS servers are used
def _is_set_to_nord_ipv6():
    return "2400:bb40:8888::100" and "2400:bb40:4444::100" in sh.cat("/etc/resolv.conf")


# returns True when Threat Protection Lite DNS servers are used
def _is_set_to_threat_protection_lite_ipv6():
    return "2400:bb40:4444::103" and "2400:bb40:8888::103" in sh.cat("/etc/resolv.conf")


# returns True when NordVPN app has not modified the DNS
def is_unset():
    return not (
        _is_set_to_nord()
        and _is_set_to_nord_ipv6()
        and _is_set_to_threat_protection_lite()
        and _is_set_to_threat_protection_lite_ipv6()
    )


def is_set_for(threat_protection_lite, ipv6="off"):
    logging.log(sh.cat("/etc/resolv.conf"))
    if threat_protection_lite == "on" and ipv6 == "on":
        return _is_set_to_threat_protection_lite_ipv6()

    if threat_protection_lite == "on":
        return _is_set_to_threat_protection_lite()

    if ipv6 == "on":
        return _is_set_to_nord_ipv6() and _is_set_to_nord()

    return _is_set_to_nord()
