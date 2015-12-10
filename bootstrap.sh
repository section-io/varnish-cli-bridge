#!/bin/bash
#set -o xtrace
set -o errexit
cd /home/vagrant

# install Varnish
apt-key list | grep varnish >/dev/null ||
  curl -s https://repo.varnish-cache.org/GPG-key.txt | sudo apt-key add - >/dev/null

test -f /etc/apt/sources.list.d/varnish-cache.list || {
  echo "deb https://repo.varnish-cache.org/ubuntu/ precise varnish-4.0" | sudo tee /etc/apt/sources.list.d/varnish-cache.list >/dev/null
  sudo apt-get update
}

command -v varnishadm >/dev/null ||
  sudo apt-get install --assume-yes varnish=3.0.5-2

#install git
command -v git >/dev/null ||
  sudo apt-get install --assume-yes git

# install Golang
golang_version=1.5.1
golang_download_file="go${golang_version}.linux-amd64.tar.gz"
golang_download_url="https://storage.googleapis.com/golang/${golang_download_file}"
#golang_source_url="https://storage.googleapis.com/golang/go${golang_version}.src.tar.gz"
golang_download_sha1=46eecd290d8803887dec718c691cc243f2175fe0

test -f "${golang_download_file}" ||
  wget --timestamping "${golang_download_url}"

echo "${golang_download_sha1} ${golang_download_file}" | sha1sum -c - >/dev/null

test -x /usr/local/go/bin/go ||
  sudo tar --extract --directory /usr/local --file "${golang_download_file}"

function linkgo {
  sudo rm -f /usr/local/bin/go
  sudo ln --symbolic /usr/local/go/bin/go /usr/local/bin/go
}

readlink -e /usr/local/bin/go >/dev/null || linkgo
# TODO re-link go if it points to the wrong one

go version

mkdir --parents /home/vagrant/go/src/github.com/section-io/
test -a /home/vagrant/go/src/github.com/section-io/varnish-cli-bridge ||
  ln --symbolic /vagrant /home/vagrant/go/src/github.com/section-io/varnish-cli-bridge

grep '^export GOPATH=' .bashrc >/dev/null || {
  echo -e '\nexport GOPATH=/home/vagrant/go' >>.bashrc
  echo 'You should execute `source ~/.bashrc` to set your GOPATH'
}

