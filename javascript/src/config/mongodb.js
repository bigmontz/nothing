import { MongoClient } from "mongodb";

export async function configureMongodbDriver(address, user, password) {
  const uri = `mongodb://${user}:${password}@${address}`
  console.log(`Connecting to Mongodb on ${uri}`)
  const mongoClient = new MongoClient(uri, { useNewUrlParser: true });
  return await mongoClient.connect();
}
