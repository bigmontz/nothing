import neo4j from 'neo4j-driver-lite'

export function configureNeo4jDriver(url, user, password) {
  console.log(`Connecting to Neo4j on ${url} -> ${user}:${password}`)
  return neo4j.driver(url, neo4j.auth.basic(user, password), {
    useBigInt: true,
    logging: neo4j.logging.console('debug')
  });
}
