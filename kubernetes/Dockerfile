FROM debian:stretch-slim
WORKDIR /work
EXPOSE 5310
RUN apt-get update
RUN apt-get install -y wget unzip
RUN wget https://github.com/cdsl-research/nickns/releases/latest/download/release-lin64.zip
RUN unzip release-lin64.zip
ENTRYPOINT ["/work/app"]
