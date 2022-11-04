FROM centos7.2-common:1.1.0

RUN mkdir /home/admin/sentinel && \
    cd /home/admin/sentinel && \
    mkdir output script
COPY redis-cli redis-sentinel docker-entrypoint.sh  check_sentinel_alive.sh /usr/local/bin/
COPY sentinel.conf /home/admin/sentinel
COPY init_sentinel.py /home/admin/sentinel/script
EXPOSE 26379

ENTRYPOINT ["docker-entrypoint.sh"]

