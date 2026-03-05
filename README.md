# podsh

This application acts as a gateway for ssh => kubectl exec -it, with user authentication and owner-id check for pods.

## Client side usage

```bash
ssh <deployment-name>@<public-ssh-host>:<public-ssh-port>
# This server validates SSH access against stored ssh public key,
# based on the public key used we keep a mapping of the users ID.
# Then the app looks up pods that correspond to <deployment-name> in the specified namespace by
# checking for `app.kubernetes.io/deploy-name` with `owner-id` matching the mapped id.
# Then a interactive shell is established using the kubernetes api mapped to the established SSH connection.
# voila a "ssh" shell for the pod, and the pod doesnt even have to have sshd installed, configured and running ;)
```
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


## TODO

[x] pre-auth rate limiting
[ ] handshake deadlines
[x] tarpitting (sleep random time after failed auth)
[ ] security log
[x] tunneling support
[x] command exec support
[x] scp / sftp support

## Architecture / integration with go-deploy

The user information is stored in the mongodb database that `go-deploy` uses. But this runs in another cluster. This data needs to be synced.
Current solution is to poll this and populate a redis cache, not the best/cleanest solution but I think it will work ok.
