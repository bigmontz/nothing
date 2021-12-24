import pg from 'pg';

export function configurePostgresDriver(host, user, password, database, port) {
  console.log(`Connecting to Postgres on ${host}`)
  const config = {
    user: user,
    password: password,
    host: host
  }

  if (port) {
    config.port = port;
  }
  
  if (database) {
    config.database = database;
  }

  return new pg.Pool(config);
}
