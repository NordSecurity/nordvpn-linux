Vagrant.configure("2") do |config|
  config.vm.provider :libvirt do |libvirt|
  end

  config.vm.network "forwarded_port", guest: 6960, host: 6969
  config.vm.network "private_network", type: "dhcp"
  
  if ENV["CUSTOM_ENV_GITLAB_CI"] != "true"
    config.vm.synced_folder "dist/app", "/vagrant", type: "nfs", nfs_version: 4
  end
  config.vm.provision :shell, inline: "sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0"

  config.vm.define "debian10", autostart: false do |debian|
    debian.vm.box = "generic/debian10"
  end

  config.vm.define "debian11", autostart: false do |debian|
    debian.vm.box = "generic/debian11"
  end

  config.vm.define "ubuntu18", autostart: false do |ubuntu|
    ubuntu.vm.box = "generic/ubuntu1804"
  end

  config.vm.define "ubuntu20", autostart: false do |ubuntu|
    ubuntu.vm.box = "generic/ubuntu2004"
  end

  config.vm.define "ubuntu22", autostart: false do |ubuntu|
    ubuntu.vm.box = "generic/ubuntu2204"
  end

  config.vm.define "fedora35", autostart: false do |fedora|
    fedora.vm.box = "generic/fedora35"
  end

  config.vm.define "fedora36", autostart: false do |fedora|
    fedora.vm.box = "generic/fedora36"
  end

  config.vm.define "default", primary: true  do |default|
    default.vm.box = "generic/debian11"
    default.vm.synced_folder "./", "/vagrant"
    default.vm.provision "shell", path: "ci/provision.sh"
    if ENV["CUSTOM_ENV_GITLAB_CI"] == "true"
      # `gitlab-runner` is required for CI to be able to download previous stage artifacts: download_artifacts
      # https://docs.gitlab.com/runner/executors/custom.html
      # Copy `gitlab-runner` from runner host to guest VM.
      default.vm.provision "file", source: `which gitlab-runner`.strip(), destination: "/home/vagrant/gitlab-runner"
      default.vm.provision "shell", inline: "cp /home/vagrant/gitlab-runner /usr/local/bin/gitlab-runner", privileged: true
    end

  end
end
