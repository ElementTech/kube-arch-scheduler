FROM debian:stable-slim

WORKDIR /

RUN apt-get update && apt-get install -y ca-certificates

COPY kube-arch-scheduler /usr/local/bin

CMD ["kube-arch-scheduler"]
