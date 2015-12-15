FROM scratch

ADD https://raw.githubusercontent.com/bagder/ca-bundle/master/ca-bundle.crt /etc/ssl/ca-bundle.pem

# ENVIRONMENTS
ENV GOENV production

WORKDIR /app
CMD ["./registry-proxy"]
ADD assets/ /app/assets
ADD views/ /app/views
ADD registry.db /app/db/
ADD registry-proxy /app/registry-proxy
