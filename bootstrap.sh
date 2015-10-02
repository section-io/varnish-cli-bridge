#!/bin/sh

#wget https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz

sudo add-apt-repository --yes ppa:evarlast/golang1.5
sudo apt-get update

sudo apt-get purge --assume-yes golang
sudo add-apt-repository --yes --remove ppa:duh/golang

sudo apt-get install --assume-yes golang=2:1.5*

sudo update-alternatives --install /usr/bin/go golang-go /usr/bin/golang-go 1

#~/.bashrc contains `export GOPATH=/home/vagrant/go`
mkdir --parents /home/vagrant/go/src/github.com/section-io/
ln --symbolic /vagrant /home/vagrant/go/src/github.com/section-io/varnish-cli-bridge

curl https://repo.varnish-cache.org/GPG-key.txt | sudo apt-key add -
echo "deb https://repo.varnish-cache.org/ubuntu/ precise varnish-4.0" | sudo tee /etc/apt/sources.list.d/varnish-cache.list
sudo apt-get update
sudo apt-get install --assume-yes varnish
