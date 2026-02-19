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

