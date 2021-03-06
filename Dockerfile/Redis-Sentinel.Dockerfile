FROM centos7.2-common:1.1.0
RUN mkdir /home/admin/sentinel && \
    cd /home/admin/sentinel && \
    mkdir output && \
    mkdir script
COPY redis-cli redis-sentinel docker-entrypoint.sh  check_sentinel_alive.sh /usr/local/bin/
COPY sentinel.conf /home/admin/sentinel
COPY init_sentinel.py /home/admin/sentinel/script
ENTRYPOINT ["docker-entrypoint.sh"]
EXPOSE 26379
