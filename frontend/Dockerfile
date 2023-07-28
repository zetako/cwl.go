FROM frolvlad/alpine-glibc:latest

WORKDIR /root/

COPY example/local /root/deployment/config/local

COPY example/pre /root/deployment/config/pre

COPY cwl.go /usr/bin/

ENTRYPOINT ["cwl.go", "serve"]

EXPOSE 4321