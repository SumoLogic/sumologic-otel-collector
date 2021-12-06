#!/usr/bin/env bash

export BUILDER_VERSION="$(grep '^BUILDER_VERSION' /sumologic/otelcolbuilder/Makefile | sed 's/BUILDER_VERSION ?= //')"
export GO_VERSION=1.17

# Install opentelemetry-collector-builder
curl -LJ \
    "https://github.com/open-telemetry/opentelemetry-collector-builder/releases/download/v${BUILDER_VERSION}/opentelemetry-collector-builder_${BUILDER_VERSION}_linux_amd64" \
    -o /usr/local/bin/opentelemetry-collector-builder \
    && chmod +x /usr/local/bin/opentelemetry-collector-builder

sudo apt update -y
sudo apt install -y \
    make \
    gcc \
    python3-pip

# Install Go
curl -LJ "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o go.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go.linux-amd64.tar.gz \
    && rm go.linux-amd64.tar.gz \
    && ln -s /usr/local/go/bin/go /usr/local/bin

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
sed -i 's#secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"#secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/opt/puppetlabs/bin"#g' /etc/sudoers

# Install chef
curl -L https://www.opscode.com/chef/install.sh | sudo bash

# accepts chef-solo licenses
chef-solo --chef-license=accept || true
su vagrant -c 'chef-solo --chef-license=accept' || true

# Install docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
apt-get install -y docker-ce docker-ce-cli containerd.io
usermod -aG docker vagrant

# Install LuaJIT
wget https://luajit.org/download/LuaJIT-2.1.0-beta3.tar.gz
tar -xzf LuaJIT-2.1.0-beta3.tar.gz
cd LuaJIT-2.1.0-beta3/
make && make install
cd src
ln -sf luajit /usr/local/bin/luajit
cp libluajit.so /usr/local/lib/libluajit.so
ldconfig
