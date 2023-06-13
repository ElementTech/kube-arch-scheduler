FROM debian:stretch-slim

WORKDIR /

COPY kube-arch-scheduler /usr/local/bin

CMD ["kube-arch-scheduler"]
