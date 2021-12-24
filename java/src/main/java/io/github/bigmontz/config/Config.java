package io.github.bigmontz.config;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
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
        String dbType = Env.getOrThrow("DB_TYPE", () -> new RuntimeException("missing DB_TYPE envvar"));
        switch (dbType) {
            case "neo4j":
                return new UserNeo4jRepository(neo4jDriver());
            default:
                throw new IllegalStateException(String.format("unsupported DB_TYPE %s", dbType));
        }
    }

    private static Driver neo4jDriver() {
        return GraphDatabase.driver(
                Env.getOrDefault("NEO4J_URL", "neo4j://localhost"),
                AuthTokens.basic(
                        Env.getOrDefault("NEO4J_USER", "neo4j"),
                        Env.getOrDefault("NEO4J_PASSWORD", "pass")));
    }
}
