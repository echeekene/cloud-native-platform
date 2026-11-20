# Pinned CoreDNS image with our config baked in for local testing.
# In Kubernetes the same files are mounted from a ConfigMap instead.
FROM coredns/coredns:1.12.1
COPY Corefile /etc/coredns/Corefile
COPY zones/ /etc/coredns/zones/
EXPOSE 1053/udp 1053/tcp 9153/tcp
ENTRYPOINT ["/coredns", "-conf", "/etc/coredns/Corefile"]
