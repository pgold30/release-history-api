FROM postgres:9.6.6

ADD docker-entrypoint-initdb.d /docker-entrypoint-initdb.d
RUN chown postgres:postgres /docker-entrypoint-initdb.d/*
