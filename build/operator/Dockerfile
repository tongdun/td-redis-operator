FROM redis:alpine3.12

RUN mkdir /app
WORKDIR /app

COPY _output/operator operator

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.12/main/" > /etc/apk/repositories

RUN apk update \
        && apk upgrade \
        && apk add --no-cache bash \
        bash-doc \
        bash-completion \
        && rm -rf /var/cache/apk/* \
        && /bin/bash

CMD ["/app/operator"]

