FROM golang:1.25.6-trixie

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        curl \
        git \
        make \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Tokyo

WORKDIR /src

CMD [ "go", "run", "./cmd/handler" ]
