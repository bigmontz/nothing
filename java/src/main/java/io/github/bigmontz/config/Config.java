package io.github.bigmontz.config;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.mongodb.client.MongoClient;
import com.mongodb.client.MongoClients;
import io.github.bigmontz.repository.UserMongoRepository;
import io.github.bigmontz.repository.UserNeo4jRepository;
import io.github.bigmontz.repository.UserRepository;
import org.neo4j.driver.AuthTokens;
import org.neo4j.driver.Driver;
import org.neo4j.driver.GraphDatabase;

import java.time.ZonedDateTime;

import static com.google.gson.FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES;

public class Config {

    public static Gson gson() {
        return new GsonBuilder()
                .setFieldNamingPolicy(LOWER_CASE_WITH_UNDERSCORES)
                .registerTypeAdapter(ZonedDateTime.class, new ZonedDateTimeGsonAdapter())
                .create();
    }

    public static UserRepository userRepository() {
        var dbType = Env.getOrThrow("DB_TYPE", () -> new RuntimeException("missing DB_TYPE envvar"));
        return switch (dbType) {
            case "neo4j" -> new UserNeo4jRepository(neo4jDriver());
            case "mongodb" -> new UserMongoRepository(mongoDriver());
            default -> throw new IllegalStateException(String.format("unsupported DB_TYPE %s", dbType));
        };
    }

    private static Driver neo4jDriver() {
        return GraphDatabase.driver(
                Env.getOrDefault("NEO4J_URL", "neo4j://localhost"),
                AuthTokens.basic(
                        Env.getOrDefault("NEO4J_USER", "neo4j"),
                        Env.getOrDefault("NEO4J_PASSWORD", "pass")));
    }

    private static MongoClient mongoDriver() {
        // TODO: add "?retryWrites=true"?
        return MongoClients.create(String.format("mongodb://%s:%s@%s",
                Env.getOrDefault("MONGODB_USER", "mongodb"),
                Env.getOrDefault("MONGODB_PASSWORD", "mongodb"),
                Env.getOrDefault("MONGODB_ADDRESS", "localhost")));
    }
}
