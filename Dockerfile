FROM debian:stretch-slim

WORKDIR /

COPY dist/kube-arch-scheduler /usr/local/bin

CMD ["kube-arch-scheduler"]
