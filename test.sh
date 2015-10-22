#!/usr/bin/env bash
#set -o xtrace DEBUG
set -o errexit

stdout_file=/tmp/test-output
stderr_file=/tmp/test-error
secret_file=/tmp/test-secret
wrong_secret_file=/tmp/test-wrong-secret

cat /proc/sys/kernel/random/uuid >$secret_file
cat /proc/sys/kernel/random/uuid >$wrong_secret_file

echo 'INIT executing `go test`'
go test github.com/section-io/varnish-cli-bridge || {
  echo 'Failed go tests.'
  exit 1
}

echo INIT installing bridge
go install github.com/section-io/varnish-cli-bridge || {
  echo 'Failed to install.'
  exit 1
}

echo INIT launching bridge
#NOTE `go test` appears to install to $GOPATH/bin
SECTION_IO_PASSWORD=P@ssw0rd $GOPATH/bin/varnish-cli-bridge \
  -listen-address :6083 \
  -secret-file "${secret_file}" \
  -api-endpoint http://httpbin.org/post \
  -username testuser >/tmp/bridge-output 2>/tmp/bridge-error &
proxy_pid=$!

sleep 2 # allow time for proxy to begin listening

function finally {
  echo DONE terminating bridge
  kill -s SIGTERM $proxy_pid
  wait
}
trap finally EXIT

echo -n TEST expects auth ...
exit_code=0 ; varnishadm -T :6083 ping >$stdout_file 2>$stderr_file || exit_code=$?
grep --quiet --fixed-strings 'Authentication required' $stderr_file && echo PASS || echo FAIL

echo -n TEST wrong auth fails ...
exit_code=0 ; varnishadm -S $wrong_secret_file -T :6083 ping >$stdout_file 2>$stderr_file || exit_code=$?
[ "${exitcode}" != "0" ] && echo PASS || echo FAIL

echo -n TEST correct auth succeeds ...
exit_code=0 ; varnishadm -S $secret_file -T :6083 ping >$stdout_file 2>$stderr_file || exit_code=$?
grep --quiet --fixed-strings 'PONG' $stdout_file && echo PASS || echo FAIL

echo -n TEST banner ...
exit_code=0 ; varnishadm -S $secret_file -T :6083 banner >$stdout_file 2>$stderr_file || exit_code=$?
grep --quiet --fixed-strings 'Varnish Cache CLI Bridge' $stdout_file && echo PASS || echo FAIL

echo -n TEST ban ...
exit_code=0 ; varnishadm -S $secret_file -T :6083 "ban req.url == /dummy" >$stdout_file 2>$stderr_file || exit_code=$?
grep --quiet --fixed-strings 'Ban forwarded' $stdout_file && echo PASS || echo FAIL
