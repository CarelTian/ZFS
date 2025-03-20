FROM gcr.io/etcd-development/etcd:v3.5.9

VOLUME ["/etcd-data"]
COPY etcd-config.yaml /etcd-config.yaml

EXPOSE 2379 2380


CMD ["etcd", "--config-file", "/etcd-config.yaml"]

