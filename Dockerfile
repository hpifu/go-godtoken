FROM centos:centos7

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" >> /etc/timezone

COPY docker/go-godtoken /var/docker/go-godtoken
RUN mkdir -p /var/docker/go-godtoken/log

EXPOSE 7060

WORKDIR /var/docker/go-godtoken
CMD [ "bin/godtoken", "-c", "configs/godtoken.json" ]
