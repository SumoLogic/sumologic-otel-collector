#!/usr/bin/env bash

export GO_VERSION="1.21.4"

ARCH="$(dpkg --print-architecture)"

sudo apt update -y
sudo apt install -y \
    make \
    gcc \
    python3-pip

# Install Go
curl -LJ "https://golang.org/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o go.linux-${ARCH}.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go.linux-${ARCH}.tar.gz \
    && rm go.linux-${ARCH}.tar.gz \
    && ln -s /usr/local/go/bin/go /usr/local/bin

# Install Node.js (for tools like linters etc.)
curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
apt-get install -y nodejs

make -C /sumologic install-markdownlint

# Install opentelemetry-collector-builder

su vagrant -c 'export PATH="${PATH}:/home/vagrant/bin"; make -C /sumologic/otelcolbuilder/ install-builder'

# Install ansible
pip3 install ansible

# Add puppet hosts
tee -a /etc/hosts << END
127.0.0.1 agent
END

# Install puppet server & puppet agent
wget https://apt.puppetlabs.com/puppet6-release-focal.deb
dpkg -i puppet6-release-focal.deb
apt-get update -y
apt-get install puppetserver puppet-agent -y

tee /etc/puppetlabs/puppet/puppet.conf << END
[server]
vardir = /opt/puppetlabs/server/data/puppetserver
logdir = /var/log/puppetlabs/puppetserver
rundir = /var/run/puppetlabs/puppetserver
pidfile = /var/run/puppetlabs/puppetserver/puppetserver.pid
codedir = /etc/puppetlabs/code

certname = sumologic-otel-collector
server = sumologic-otel-collector

[agent]
certname = agent
server = sumologic-otel-collector
END

# Start puppet server
systemctl start puppetserver
systemctl enable puppetserver

# Start puppet agent
systemctl start puppet
systemctl enable puppet

echo 'PATH="$PATH:/opt/puppetlabs/bin/"' >> /etc/profile
echo 'PATH="$PATH:/home/vagrant/bin:/home/vagrant/go/bin"' >> /home/vagrant/.bashrc
sed -i 's#secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"#secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/opt/puppetlabs/bin"#g' /etc/sudoers

# Install chef
curl -L https://omnitruck.chef.io/install.sh | sudo bash

# accepts chef-solo licenses
chef-solo --chef-license=accept || true
su vagrant -c 'chef-solo --chef-license=accept' || true

# Install docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository \
   "deb [arch=${ARCH}] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
apt-get install -y docker-ce docker-ce-cli containerd.io
usermod -aG docker vagrant
