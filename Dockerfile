FROM box-registry.jfrog.io/jenkins/box-centos7

LABEL com.box.name="node-alert-worker"

# Required for systemd related things to work
ENV container=docker

ADD ./build/node-alert-worker /node-alert-worker
RUN chown container:container /node-alert-worker