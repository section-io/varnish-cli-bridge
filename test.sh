#!/usr/bin/env bash
set -o xtrace
set -o errexit

stdout_file=/tmp/test-output
stderr_file=/tmp/test-error
secret_file=/tmp/test-secret

cat /proc/sys/kernel/random/uuid >$secret_file

SECTION_IO_PASSWORD=P@ssw0rd $GOPATH/bin/varnish-cli-bridge \
  -listen-address :6083 \
  -secret-file "${secret_file}" \
  -api-endpoint http://httpbin.org/post \
  -username testuser &
proxy_pid=$!

sleep 2 # allow time for proxy to begin listening

echo TEST needs auth
exit_code=0 ; varnishadm -T :6083 ping >$stdout_file 2>$stderr_file || exit_code=$?
grep --fixed 'Authentication required' $stderr_file || echo FAILED

echo TEST needs auth
exit_code=0 ; varnishadm -S $secret_file -T :6083 ping >$stdout_file 2>$stderr_file || exit_code=$?
echo $exit_code
cat $stdout_file $stderr_file

echo TEST ban
exit_code=0 ; varnishadm -S $secret_file -T :6083 "ban req.url == /dummy" >$stdout_file 2>$stderr_file || exit_code=$?
echo $exit_code
cat $stdout_file $stderr_file

echo TEST banner
exit_code=0 ; varnishadm -S $secret_file -T :6083 banner >$stdout_file 2>$stderr_file || exit_code=$?
echo $exit_code
cat $stdout_file $stderr_file

kill -s SIGTERM $proxy_pid
wait
