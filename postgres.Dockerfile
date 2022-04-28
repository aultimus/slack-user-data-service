FROM postgres:9.6
ENV POSTGRES_DB postgres
COPY schema.sql /docker-entrypoint-initdb.d/
