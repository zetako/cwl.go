FROM frolvlad/alpine-glibc:latest

WORKDIR /root/

COPY cwl.go /usr/bin/

COPY example/* /root/deployment/config/local/

CMD ["cwl.go", "serve"]

EXPOSE 4321