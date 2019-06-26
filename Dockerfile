FROM box-registry.jfrog.io/jenkins/box-centos7

LABEL com.box.name="node-alert-worker"

# Required for systemd related things to work
ENV container=docker

ADD ./ansible /ansible
RUN yum --disablerepo=packages-box install -y python-pyasn1 && \ 
    yum install -y python-pip PyYAML python-jinja2 python-httplib2 python-keyczar python-paramiko
RUN pip install -r ansible/requirements.txt
ADD ./build/node-alert-worker /node-alert-worker
ADD ./ansible /ansible
ADD config /config
RUN chown -R container:container /config  && chown -R container:container /ansible  && chown container:container /node-alert-worker