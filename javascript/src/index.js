import Server from './config/server.js';
import UserNeo4jRepository from './repository/user.neo4j.js';
import UserPostgresRepository from './repository/user.postgres.js';
import UserController from './controller/user.js';
import { createCrudRouteFor } from './routes/crud.route.js';
import { configureNeo4jDriver } from './config/neo4j.js';
import { configurePostgresDriver } from './config/postgres.js';
import { configureMongodbDriver } from './config/mongodb.js';
import UserMongodbRepository from './repository/user.mongodb.js';

const databaseAccess = await configureDatabaseAccess();

const userController = new UserController(databaseAccess.repositories.user);
const userRoute = createCrudRouteFor(userController);

const server = new Server()
server.defineRoute('/user', userRoute);
server.start(process.env.PORT || 3000)

process.on('SIGINT', async () => {
  server.stop();
  await databaseAccess.close();
  process.exit();
})


async function configureDatabaseAccess() {
  switch (process.env.DB_TYPE) {
    case 'postgres':
      return await configurePostgresDatabaseAccess();
    case 'mongodb':
      return await configureMongodbDatabaseAccess();
    case 'cockroachdb':
      return await configureCockroachdbDatabaseAccess();
    case 'neo4j':
    default:
      return await configureNeo4jDatabaseAccess();
  }
}

async function configureMongodbDatabaseAccess() {
  const driver = await configureMongodbDriver(
    process.env.MONGODB_ADDRESS || "localhost",
    process.env.MONGODB_USER || "mongodb",
    process.env.MONGODB_PASSWORD || "mongodb"
  );
  return {
    close: () => driver.close(),
    repositories: {
      user: new UserMongodbRepository(driver)
    }
  }
}

async function configurePostgresDatabaseAccess() {
  const driver = configurePostgresDriver(
    process.env.POSTGRES_URL || "localhost",
    process.env.POSTGRES_USER || "postgres",
    process.env.POSTGRES_PASSWORD || "postgres");
  return {
    close: () => driver.end(),
    repositories: {
      user: new UserPostgresRepository(driver)
    }
  }
}

async function configureCockroachdbDatabaseAccess() {
  const driver = configurePostgresDriver(
    process.env.COCKROACH_URL || "localhost",
    process.env.COCKROACH_USER || "admin",
    process.env.COCKROACH_PASSWORD || "cockroach",
    process.env.COCKROACH_DATABASE || "postgres",
    process.env.COCKROACH_PORT || 26257);
    
  driver.query("CREATE TABLE IF NOT EXISTS users (" +
	  "id SERIAL PRIMARY KEY," +
	  "username VARCHAR(255) NOT NULL," +
    "name VARCHAR(255) NOT NULL," +
    "surname VARCHAR(255) NOT NULL," +
    "password VARCHAR(255) NOT NULL," +
    "age INTEGER NOT NULL," +
    "created_at TIMESTAMP NOT NULL DEFAULT NOW(), " +
    "updated_at TIMESTAMP NOT NULL DEFAULT NOW()" +
	");");
  return {
    close: () => driver.end(),
    repositories: {
      user: new UserPostgresRepository(driver)
    }
  }
}


async function configureNeo4jDatabaseAccess() {
  const driver = configureNeo4jDriver(
    process.env.NEO4J_URL || "neo4j://localhost",
    process.env.NEO4J_USER || "neo4j",
    process.env.NEO4J_PASSWORD || "pass");
  return {
    close: () => driver.close(),
    repositories: {
      user: new UserNeo4jRepository(driver)
    }
  }
}
