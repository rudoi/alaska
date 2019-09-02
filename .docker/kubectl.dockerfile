FROM alpine

MAINTAINER Andrew Rudoi <andrewrudoi@gmail.com>

ENV KUBERNETES_VERSION="v1.15.2"

RUN apk add --update curl \
    && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && apk del --purge curl \
    && rm /var/cache/apk/*

ENTRYPOINT ["kubectl"]
CMD ["help"]
