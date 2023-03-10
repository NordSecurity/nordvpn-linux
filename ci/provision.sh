#!/bin/bash
set -euxo pipefail

# Install docker
sudo apt-get update --allow-releaseinfo-change
sudo apt-get -y install apt-transport-https ca-certificates curl gnupg lsb-release
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [trusted=yes] https://download.docker.com/linux/debian  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get -y install docker-ce docker-ce-cli containerd.io docker-compose-plugin git
sudo usermod -aG docker vagrant
docker --version

# Fix docker login command; https://stackoverflow.com/questions/50151833/cannot-login-to-docker-account
sudo apt-get -y install gnupg2 pass

# Other dependencies
sudo apt-get -y install python3-pip