#!/bin/bash

echo "setup schema to ${DBNAME}"
SQL_DATABASE="${DBNAME}" temporal-sql-tool \
  --ep mysql \
  -p 3306 \
  -u "${MYSQL_USER}" \
  -pw "${MYSQL_PWD}" \
  --pl mysql8 \
  --db "${DBNAME}" \
  setup-schema -v 0.0

echo "update schema to ${DBNAME}"
SQL_DATABASE="${DBNAME}" temporal-sql-tool \
  --ep mysql \
  -p 3306 \
  -u "${MYSQL_USER}" \
  -pw "${MYSQL_PWD}" \
  --pl mysql8 \
  --db "${DBNAME}" \
  update-schema -d ./schema/mysql/v8/temporal/versioned/

echo "setup schema to ${VISIBILITY_DBNAME}"
SQL_DATABASE="${VISIBILITY_DBNAME}" temporal-sql-tool \
  --ep mysql \
  -p 3306 \
  -u "${MYSQL_USER}" \
  -pw "${MYSQL_PWD}" \
  --pl mysql8 \
  --db "${VISIBILITY_DBNAME}" \
  setup-schema -v 0.0

echo "update schema to ${VISIBILITY_DBNAME}"
SQL_DATABASE="${VISIBILITY_DBNAME}" temporal-sql-tool \
  --ep mysql \
  -p 3306 \
  -u "${MYSQL_USER}" \
  -pw "${MYSQL_PWD}" \
  --pl mysql8 \
  --db "${VISIBILITY_DBNAME}" \
  update-schema -d ./schema/mysql/v8/visibility/versioned/
