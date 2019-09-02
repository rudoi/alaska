FROM alpine

RUN apk add --update curl \
    && curl -L https://get.helm.sh/helm-v3.0.0-beta.2-linux-amd64.tar.gz -o helm.tar.gz \
    && tar -xzvf helm.tar.gz \
    && mv linux-amd64/helm /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm \
    && rm -rf ./*.tar.gz linux-amd64 \
    && apk del --purge curl \
    && rm /var/cache/apk/*

ENTRYPOINT ["helm"]
CMD ["help"]
