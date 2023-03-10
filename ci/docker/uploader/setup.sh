#!/bin/bash
git clone -b 2.21-release https://github.com/pulp/pulp.git
git clone -b 1.10-release https://github.com/pulp/pulp_deb.git
git clone -b 2.21-release https://github.com/pulp/pulp_rpm.git

# Install pulp-admin with deb and rpm extentions
pushd /build/pulp/client_admin/ && python setup.py install && popd || exit
pushd /build/pulp/bindings/ && python setup.py install && popd || exit
pushd /build/pulp/client_lib/ && python setup.py install && popd || exit
pushd /build/pulp/common/ && python setup.py install && popd || exit
pushd /build/pulp_deb/extensions_admin/ && python setup.py install && popd || exit
pushd /build/pulp_deb/common/ && python setup.py install && popd || exit
pushd /build/pulp_rpm/extensions_admin/ && python setup.py install && popd || exit
pushd /build/pulp_rpm/common/ && python setup.py install && popd || exit
