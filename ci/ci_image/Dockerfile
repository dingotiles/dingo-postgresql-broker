FROM concourse/concourse-ci
# https://github.com/concourse/concourse/blob/master/ci/dockerfiles/concourse-ci/Dockerfile

RUN apt-get update && \
    apt-get install curl -y && \
    apt-get clean

# Install Go
RUN \
  mkdir -p /goroot && \
  curl https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz | tar xvzf - -C /goroot --strip-components=1

# Set environment variables.
ENV GOROOT /goroot
ENV GOPATH /gopath
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

# Install ETCD
RUN curl -L https://github.com/coreos/etcd/releases/download/v2.3.4/etcd-v2.3.4-linux-amd64.tar.gz -o /tmp/etcd-v2.3.4-linux-amd64.tar.gz \
      && tar xzvf /tmp/etcd-v2.3.4-linux-amd64.tar.gz -C /tmp \
      && mv /tmp/etcd-v2.3.4-linux-amd64/etcd /usr/local/bin \
      && rm -rf /tmp/etcd*
