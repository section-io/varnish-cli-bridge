#!/usr/bin/env bash
set -o xtrace
set -o errexit

stdout_file=/tmp/test-output
stderr_file=/tmp/test-error
secret_file=/tmp/test-secret

cat /proc/sys/kernel/random/uuid >$secret_file

$GOPATH/bin/varnish-cli-bridge -listen-address :6083 &
proxy_pid=$!

sleep 2 # allow time for proxy to begin listening

echo TEST needs auth
exit_code=0 ; varnishadm -T :6083 ping >$stdout_file 2>$stderr_file || exit_code=$?
grep --fixed 'Authentication required' $stderr_file || echo FAILED

echo TEST needs auth
exit_code=0 ; varnishadm -S $secret_file -T :6083 ping >$stdout_file 2>$stderr_file || exit_code=$?
echo $exit_code
cat $stdout_file $stderr_file

kill -s SIGTERM $proxy_pid
wait
