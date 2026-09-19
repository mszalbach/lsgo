FROM scratch
ARG TARGETPLATFORM

WORKDIR /app

COPY $TARGETPLATFORM/lsgo /app/lsgo

# non-root user with GID 0 for OpenShift compatibility
USER 7777:0

EXPOSE 8080
ENTRYPOINT ["/app/lsgo"]
CMD ["--help"]