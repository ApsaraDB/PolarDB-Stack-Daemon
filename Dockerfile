# Build the manager binary
FROM reg.docker.alibaba-inc.com/polar-gocommon/golang:1.13.10 as builder

ARG ssh_prv_key
ARG ssh_pub_key

ARG DIR=/go/src/gitlab.alibaba-inc.com/rds/polarstack-daemon
WORKDIR $DIR

# Add the keys and set permissions
RUN mkdir -p /root/.ssh && chmod 0700 /root/.ssh
RUN echo "$ssh_prv_key" | tr -d '\r' > /root/.ssh/id_rsa && \
    echo "$ssh_pub_key" | tr -d '\r' > /root/.ssh/id_rsa.pub && \
    chmod 600 /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa.pub

RUN echo "Host gitlab.alibaba-inc.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config

# Copy the go source
COPY polar-controller-manager polar-controller-manager
COPY cmd cmd
COPY version version
COPY go.mod go.sum ./
#
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV GOPRIVATE=gitlab.alibaba-inc.com
RUN git config --global url."git@gitlab.alibaba-inc.com:".insteadOf "https://gitlab.alibaba-inc.com/"
RUN go mod download

# Build
RUN go build -o polarstack-daemon $DIR/cmd/daemon
# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM centos:7.4.1708
FROM reg.docker.alibaba-inc.com/polarbox/alpine:3.9
LABEL maintainers="developer"


ARG APK_MIRROR=mirrors.aliyun.com
ARG CodeSource=
ARG CodeBranches=
ARG CodeVersion=

ENV CODE_SOURCE=$CodeSource
ENV CODE_BRANCHES=$CodeBranches
ENV CODE_VERSION=$CodeVersion

RUN mkdir -p /etc/ssh/ && \
     echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config && \
     echo "UserKnownHostsFile /dev/null"  >> /etc/ssh/ssh_config && \
     echo "ServerAliveInterval 3"  >> /etc/ssh/ssh_config && \
     echo "ServerAliveCountMax 2"  >> /etc/ssh/ssh_config && \
     echo "hosts: files dns" > /etc/nsswitch.conf

LABEL CodeSource=$CodeSource CodeBranches=$CodeBranches CodeVersion=$CodeVersion
#     apk del *

# RUN yum install -y http://yum.tbsite.net/alios/7/os/x86_64/Packages/openssh-7.4p1-16.alios7.x86_64.rpm && \
#     yum install -y http://yum.tbsite.net/alios/7/os/x86_64/Packages/openssh-clients-7.4p1-16.alios7.x86_64.rpm && \
#     echo  "StrictHostKeyChecking no" >> /etc/ssh/ssh_config && \
#     echo "UserKnownHostsFile /dev/null"  >> /etc/ssh/ssh_config && \
#     echo "ServerAliveInterval 3"  >> /etc/ssh/ssh_config && \
#     echo "ServerAliveCountMax 2"  >> /etc/ssh/ssh_config && \
#     yum clean --enablerepo=* all && \
#     touch /var/lib/rpm/*

#RUN yum install -y openssh-clients -b current && \
#    echo  "StrictHostKeyChecking no" >> /etc/ssh/ssh_config && \
#    yum clean --enablerepo=* all && \
#    touch /var/lib/rpm/*
WORKDIR /bin/
CMD [ "polarstack-daemon" ]


COPY --from=builder /go/src/gitlab.alibaba-inc.com/rds/polarstack-daemon/polarstack-daemon /usr/local/bin/

#RUN chmod +x /usr/local/bin/cloudprovider
#RUN apk add --no-cache \
#        libc6-compat
