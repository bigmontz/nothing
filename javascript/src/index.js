import Server from './config/server.js';
import UserNeo4jRepository from './repository/user.neo4j.js';
import UserController from './controller/user.js';
import { createCrudRouteFor } from './routes/crud.route.js';
import { configureNeo4jDriver } from './config/neo4j.js';

const driver = configureNeo4jDriver(
  process.env.NEO4J_URL|| 'neo4j://127.0.0.1:7687', 
  process.env.NEO4J_USER || 'neo4j',
  process.env.NEO4J_PASSWORD || 'pass');

const userRepository = new UserNeo4jRepository(driver);

const userController = new UserController(userRepository);
const userRoute = createCrudRouteFor(userController);

const server = new Server()
server.defineRoute('/user', userRoute);
server.start(process.env.PORT || 3000)

process.on('SIGINT', () => {
  server.stop()
  process.exit()
})
