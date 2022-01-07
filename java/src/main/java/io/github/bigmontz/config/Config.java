package io.github.bigmontz.config;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.mongodb.client.MongoClient;
import com.mongodb.client.MongoClients;
import io.github.bigmontz.repository.UserCockroachRepository;
import io.github.bigmontz.repository.UserMongoRepository;
import io.github.bigmontz.repository.UserNeo4jRepository;
import io.github.bigmontz.repository.UserPostgresRepository;
import io.github.bigmontz.repository.UserRepository;
import org.neo4j.driver.AuthTokens;
import org.neo4j.driver.Driver;
import org.neo4j.driver.GraphDatabase;
import org.postgresql.ds.PGSimpleDataSource;

import javax.sql.DataSource;
import java.sql.SQLException;
import java.time.ZonedDateTime;

import static com.google.gson.FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES;

public class Config {

    public static Gson gson() {
        return new GsonBuilder()
                .setFieldNamingPolicy(LOWER_CASE_WITH_UNDERSCORES)
                .registerTypeAdapter(ZonedDateTime.class, new ZonedDateTimeGsonAdapter())
                .create();
    }

    public static UserRepository<?> userRepository() {
        var dbType = Env.getOrThrow("DB_TYPE", () -> new RuntimeException("missing DB_TYPE envvar"));
        return switch (dbType) {
            case "neo4j" -> new UserNeo4jRepository(neo4jDriver());
            case "mongodb" -> new UserMongoRepository(mongoDriver());
            case "postgres" -> new UserPostgresRepository(postgresDriver());
            case "cockroachdb" -> {
                DataSource dataSource = cockroachDriver();
                createUserTable(dataSource);
                yield new UserCockroachRepository(dataSource);
            }
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

    private static DataSource postgresDriver() {
        var url = String.format("jdbc:postgresql://%s/", Env.getOrDefault("POSTGRES_URL", "localhost"));
        PGSimpleDataSource dataSource = new PGSimpleDataSource();
        dataSource.setURL(url);
        dataSource.setUser(Env.getOrDefault("POSTGRES_USER", "postgres"));
        dataSource.setPassword(Env.getOrDefault("POSTGRES_PASSWORD", "postgres"));
        return dataSource;
    }

    private static DataSource cockroachDriver() {
        PGSimpleDataSource dataSource = new PGSimpleDataSource();
        dataSource.setServerNames(new String[]{Env.getOrDefault("COCKROACH_URL", "localhost")});
        dataSource.setUser(Env.getOrDefault("COCKROACH_USER", "admin"));
        dataSource.setPassword(Env.getOrDefault("COCKROACH_PASSWORD", "cockroach"));
        dataSource.setDatabaseName(Env.getOrDefault("COCKROACH_DATABASE", "postgres"));
        dataSource.setSsl(false);
        dataSource.setPortNumbers(new int[]{Env.getOrDefault("COCKROACH_PORT", 26257, Integer::parseInt)});
        return dataSource;
    }

    private static void createUserTable(DataSource dataSource) {
        try (var connection = dataSource.getConnection();
             var statement = connection.createStatement()) {

            statement.execute("""
                    CREATE TABLE IF NOT EXISTS users (
                        id SERIAL PRIMARY KEY,
                        username VARCHAR(255) NOT NULL,
                        name VARCHAR(255) NOT NULL,
                        surname VARCHAR(255) NOT NULL,
                        password VARCHAR(255) NOT NULL,
                        age INTEGER NOT NULL,
                        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
                    );""");
        } catch (SQLException e) {
            throw new RuntimeException(e);
        }
    }
}
