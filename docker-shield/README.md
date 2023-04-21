# Docker Shield

Docker Shield is a plugin for Docker that prevents privileged mode when starting containers and using exec to enter running containers.  It also prevents users from modifying the security profiles for AppArmor and Seccomp.

# Installation

To build and enable the plugin:
```
make plugin
```

Then to use the plugin, open the docker.service file, usually located at /lib/systemd/system/docker.service, and add `--authorization-plugin=docker-shield` to the `ExecStart` command.  Here is an example of what that may look like:
```
ExecStart=/usr/bin/dockerd --authorization-plugin=docker-shield --debug -H fd:// --containerd=/run/containerd/containerd.sock
```

The reload and restart the service:
```
sudo systemctl daemon-reload
sudo systemctl restart docker
```

# Testing

We can test that the plugin is installed and working properly by passing the `--privileged` option to `docker run`.
```
$ docker run --privileged alpine sh
```

And we should get the following error output:
```
docker: Error response from daemon: authorization denied by plugin docker-shield:latest: Privileged containers not allowed.
See 'docker run --help'.
```
