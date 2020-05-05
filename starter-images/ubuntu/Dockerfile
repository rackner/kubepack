ARG OS_VERSION
FROM ubuntu:${OS_VERSION}
ARG KUBE_VERSION
RUN apt-get update && apt-get install -y apt-transport-https curl
RUN apt-get update && apt-get install -y gnupg2
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN touch /etc/apt/sources.list.d/kubernetes.list 
RUN echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN apt-get install -qy docker.io
RUN apt-get install -qy kubelet=$KUBE_VERSION
RUN apt-get install -qy kubectl=$KUBE_VERSION
RUN apt-get install -qy kubeadm=$KUBE_VERSION