FROM frolvlad/alpine-glibc:latest

WORKDIR /root/

COPY example/* /root/deployment/config/local/

COPY cwl.go /usr/bin/

ENTRYPOINT ["cwl.go", "serve"]

EXPOSE 4321