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

## TODO

pre-auth rate limiting
handshake deadlines
tarpitting (sleep random time after failed auth)
security log
tunneling support
command exec support

## Architecture / integration with go-deploy

The user information is stored in the mongodb database that `go-deploy` uses. But this runs in another cluster. This data needs to be synced.

go deploy should publish changes to nats

BELOW IS OLD DESIGN!

### identity-event-gateway

`identity-event-gateway` is a service that runs on the `local` cluster (where go-deploy runs), subscribes to events on the user collection in the db. On event relays it to `gRPC` clients that have subscribed. `gRPC` clients authorize through mTLS.

### identity-projection-sync

`identity-projection-sync` is a service that runs where podsh is deployed. It acts as a client that subscribes to the `identity-event-gateway` and populates a redis cache with the user data, (pk => user info).

