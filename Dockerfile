FROM scratch
COPY yeet /
EXPOSE 8080
ENTRYPOINT ["/yeet"]