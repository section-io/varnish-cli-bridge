FROM centurylink/ca-certs
# ^^ ca-certs is scratch plus trusted CAs

MAINTAINER Jason Stangroome <jason@section.io>

ADD varnish-cli-bridge /varnish-cli-bridge

# setup volume map for secret file and maybe VCL files in the future
VOLUME /etc/varnish

# assume default port
#ENV VARNISH_CLI_BRIDGE_LISTEN_ADDRESS :6082
EXPOSE 6082

ENTRYPOINT ["/varnish-cli-bridge"]
