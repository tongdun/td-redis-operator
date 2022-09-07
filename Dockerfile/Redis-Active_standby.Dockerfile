FROM centos7.2-common:1.1.0

RUN mkdir /home/admin/redis && \
    cd /home/admin/redis && \
    mkdir output data script
COPY redis-server redis-cli redis-check-rdb docker-entrypoint.sh check_ok.py  check_master.py  redis-login.sh /usr/local/bin/
COPY init_redis.py /home/admin/redis/script
EXPOSE 6379

ENTRYPOINT ["docker-entrypoint.sh"]

