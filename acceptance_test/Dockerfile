FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y \
        busybox \
        ca-certificates \
        dumb-init \
        golang \
        grep \
        gzip \
        initramfs-tools \
        lz4 \
        sed \
        sudo \
        zstd 

RUN sed -i 's/COMPRESS=.*/COMPRESS=lz4/g'  /etc/initramfs-tools/initramfs.conf \
    && echo "COMPRESSLEVEL=0" >> /etc/initramfs-tools/initramfs.conf 

RUN apt-get install -y \
        linux-image-amd64 

# Set the user to the current user
ARG USER_ID
ARG GROUP_ID

RUN groupadd -g $GROUP_ID myuser && \
    useradd -u $USER_ID -g $GROUP_ID -s /bin/bash myuser && \
    mkdir -p /home/myuser && \
    chown -R myuser:myuser /home/myuser && \
    echo "myuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER myuser
WORKDIR /home/myuser

COPY *.sh /

CMD ["dumb-init", "bash", "/entrypoint.sh"]
