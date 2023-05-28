FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y \
        busybox \
        ca-certificates \
        dumb-init \
        gcc \
        gcc-aarch64-linux-gnu \
        gcc-arm-linux-gnueabihf \
        golang \
        grep \
        gzip \
        initramfs-tools \
        linux-image-amd64 \
        lz4 \
        p7zip-full \
        sed \
        sudo \
        udhcpc \
        zstd 

RUN sed -i 's/COMPRESS=.*/COMPRESS=lz4/g'  /etc/initramfs-tools/initramfs.conf \
    && echo "COMPRESSLEVEL=0" >> /etc/initramfs-tools/initramfs.conf 

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

ENTRYPOINT ["dumb-init", "bash"]
