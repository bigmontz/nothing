import { MongoClient } from "mongodb";

export async function configureMongdbDriver(address, user, password) {
  const uri = `mongodb://${user}:${password}@${address}`
  console.log(`Connecting to Mongodb on ${uri}`)
  const mongoclient = new MongoClient(uri, { useNewUrlParser: true });
  return await mongoclient.connect();
}
