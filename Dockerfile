FROM debian:stretch-slim

WORKDIR /

COPY _output/bin/kube-arch-scheduler /usr/local/bin

CMD ["kube-arch-scheduler"]
