import pg from 'pg';

export function configurePostgresDriver(host, user, password) {
  console.log(`Connecting to Postgres on ${host}`)
  return new pg.Pool({
    user: user,
    password: password,
    host: host,
  });
}
