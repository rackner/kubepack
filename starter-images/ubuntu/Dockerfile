FROM ubuntu:${OS_VERSION}
COPY . /app
RUN make /app
CMD python /app/app.py
RUN apt-get update && apt-get install -y apt-transport-https curl
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
RUN cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list && \
deb https://apt.kubernetes.io/ kubernetes-xenial main && \
EOF
RUN apt-get update
RUN apt-get install -qy docker.io
RUN apt-get install -qy kubelet=${KUBE_VERSION}
RUN apt-get install -qy kubectl=${KUBE_VERSION}
RUN apt-get install -qy kubeadm=${KUBE_VERSION}