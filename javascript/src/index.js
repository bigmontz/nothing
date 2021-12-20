import Server from './config/server.js';
import UserNeo4jRepository from './repository/user.neo4j.js';
import UserPostgresRepository from './repository/user.postgres.js';
import UserController from './controller/user.js';
import { createCrudRouteFor } from './routes/crud.route.js';
import { configureNeo4jDriver } from './config/neo4j.js';
import { configurePostgresDriver } from './config/postgres.js';

const databaseAccess = configureDatabaseAccess();

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


function configureDatabaseAccess() {
  switch (process.env.DB_TYPE) {
    case 'postgres':
      return configurePostgresDatabaseAccess();
    case 'neo4j':
    default:
      return configureNeo4jDatabaseAccess();
  }
}

function configurePostgresDatabaseAccess() {
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

function configureNeo4jDatabaseAccess() {
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
