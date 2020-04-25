# kubepack

Extremely simple Kubernetes and application packaging using KubeAdm for minimal laboratory, test-bench or air-gapped environments.

## Use Case

You want to package container images, application manifests, install scripts and KubeAdm components into a single tarball to be distributed to air gapped environments. You don't want to use non-vanilla or Kubernetes-in-Docker options like Gravitational or Rancher (great tools, don't get me wrong). You want upstream Kubeadm for a specific version, and nothing more or less.

This project currently does not have any concept of plugins. To add new applications to your bundle, all you need to do is add YAML (Helm support on the way) manifests to the apps folder that you pass to `kubepack`.

## Non-Goals

- Robust kubernetes deployment infrastructure for cloud-based solutions : There is probably a much better tool for this, or any case where workloads can easily access the internet.
- Opinionated take on how Kubernetes should be deployed : Kubepack supports upstream KubeAdm, but it doesn't run any installation commands for you.
- A Kubernetes Distribution
- CI/CD

## Prerequisites

- Docker
- Linux (for now)
- Machine with internet access to do the initial pull of the images.

## Bundle Creation

Kubepack is a very simple tool. It only cares about two things, a YAML manifest with versioning information for your Kubernetes components and target runtime, and a folder with all your application manifests (currently supporting normal k8s yaml files, though will support Helm charts eventually).

All you need to do is fork or clone the repository, replace the values in the config.yaml with your targeted versions and OS, and run one command:

```
kubepack pack --apps=/path/to/manifest/folder --output=cluster.tar
```

That's it. Kubepack will spin up a docker container with your target runtime (support for BYO container will be added later), pull all images needed in Kubernetes installation and the manifests, and bundle them in a tarball along with anything else (scripts, etc) that you put in the manifest folder.

## Installation

Once you get the tarball onto the target machines (flash drive, scp, punch cards, choose your own adventure here), installation is simple.

Kubepack does not mess with the installation process in any way. Once you run the `unpack` command, use Kubeadm the way you normally would.

Regardless of whether a node is a worker or a master, run the following.

```
tar -xvf cluster.tar
sudo ./kubepack unpack
```

This will get your environment setup with the pre-packaged kubernetes components, including the version of Docker you specified in the manifest, and all the images you need for a full install of all your apps. After that, it's business as usual, which can of course be automated as you wish.

```
kubeadm init...
```

and

```
kubeadm join...
```

If you included an application install script or other automation in your apps folder, it can be called from the master. For an example install script, check the examples folder. If this tool has enough adoption, it may make sense to add more robust support of install scripts.
