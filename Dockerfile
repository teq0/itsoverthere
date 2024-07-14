FROM scratch
ARG ARCH
COPY ./build/$ARCH/itsoverthere /itsoverthere
ENTRYPOINT ["/itsoverthere"]
CMD []
# EOF

