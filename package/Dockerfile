FROM ubuntu:20.04
RUN apt-get update && apt-get install --no-install-recommends -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY bin/spot-manager /usr/local/bin/spot-manager
CMD ["spot-manager"]