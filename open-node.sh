#!/usr/bin/env sh

#docker run -it --rm -v "$(pwd):/src" -u "$(id -u):$(id -g)" --network host --workdir /src/webui node:20 /bin/bash
docker run -it --rm -v "$(pwd):/src" -u "root:root" --network host --workdir /src/webui node:20 /bin/bash
