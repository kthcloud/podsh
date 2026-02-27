docker run \
  --rm \
  -it \
  --read-only \
  --cap-drop ALL \
  --cap-add NET_BIND_SERVICE \
  --security-opt no-new-privileges \
  --pids-limit 11 \
  --memory 128m \
  --tmpfs /tmp:rw,noexec,nosuid,size=16m \
  podsh-1

