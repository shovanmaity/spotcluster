FROM alpine:3.9 as cert
RUN apk --update add ca-certificates

FROM scratch
ENV PATH=/bin
COPY bin/spot-manager /bin/spot-manager
COPY --from=cert /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD ["spot-manager"]
