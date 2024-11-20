import sh


def setup_module(module):  # noqa: ARG001
    sh.sudo.apt.purge("-y", "nordvpn")
