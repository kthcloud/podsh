# podsh

`podsh` lets you get a shell for pods in k8s using ssh, allowing you to use containers for development. It acts as a SSH gateway that authentates and authorizes pod access based on SSH keys added to your kthcloud account.

In kthcloud this enables development on pods that have access to shared GPUs using DRA + MPS, perfect for ML development.

<img width="2560" height="1440" alt="image" src="https://github.com/user-attachments/assets/7a3c6f6f-7765-4630-96ca-7f248256daeb" />

The picture above shows it in action with vscode remote ssh (to the left) and a normal ssh session (to the right).

## Client side usage

## Normal SSH

```bash
ssh <deployment-name>@<public-ssh-host>:<public-ssh-port>
# This server validates SSH access against stored ssh public key,
# based on the public key used we keep a mapping of the users ID.
# Then the app looks up pods that correspond to <deployment-name> in the specified namespace by
# checking for `app.kubernetes.io/deploy-name` with `owner-id` matching the mapped id.
# Then a interactive shell is established using the kubernetes api mapped to the established SSH connection.
# voila a "ssh" shell for the pod, and the pod doesnt even have to have sshd installed, configured and running ;)
```

## Using vscode remote SSH

Due to a bug in the `vscode-remote-development` extension pack, remote SSH in vscode only works if `devcontainers` are disabled, since it tries to bootstrap the devcontainers extension and fails to do this, due to getting stuck probing the container.

## Local dev

### Prerequisites

- mise
- docker
- docker buildx (for bake support)
- bash (for the mise task scripts)

### Quick start

In the repo run:

```bash
mise run deploy-dev
```

> [!NOTE]
> If the command above fails the first time, try re-running it. (TODO: fix)

You can check the status of the pods by running the command below

```bash
kubectl get pods -n podsh
```

When all are ready you can run the command below to try getting a shell over ssh for a mock deployment, with a included dev ssh-key pair in this repo by running:

```bash
mise run dev-example
```


## Features

- [x] pre-auth rate limiting
- [x] handshake deadlines
- [x] tarpitting
- [x] tunneling support
- [x] command exec support
- [x] scp / sftp support
- [x] prometheus metrics
- [x] horizontally scalable

## Architecture / integration with go-deploy

The user information is stored in the mongodb database that `go-deploy` uses. But this runs in another cluster. This data needs to be synced.
Current solution is to poll this and populate a redis cache, not the best/cleanest solution but I think it will work ok.

