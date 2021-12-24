package io.github.bigmontz.repository;

import org.neo4j.driver.Driver;
import org.neo4j.driver.Record;
import org.neo4j.driver.Result;
import org.neo4j.driver.Session;
import org.neo4j.driver.TransactionWork;
import org.neo4j.driver.types.Node;

import java.time.ZoneId;
import java.time.ZonedDateTime;
import java.util.Map;
import java.util.Optional;

public class UserNeo4jRepository implements UserRepository {

    private final Driver driver;

    public UserNeo4jRepository(Driver driver) {
        this.driver = driver;
    }

    @Override
    public User create(User user) {
        try (Session session = driver.session()) {
            return session.writeTransaction(userInsertion(user));
        }
    }

    @Override
    public Optional<User> findById(Object userId) {
        try (Session session = driver.session()) {
            return session.readTransaction(userRetrieval((int)userId));
        }
    }

    @Override
    public boolean updatePassword(Object userId, PasswordUpdate passwordUpdate) {
        try (Session session = driver.session()) {
            return session.writeTransaction(userPasswordUpdate((int)userId, passwordUpdate));
        }
    }

    @Override
    public void close() {
        driver.close();
    }

    private TransactionWork<User> userInsertion(User user) {
        return tx -> {
            Result result = tx.run("""               
                    CREATE (user:User {
                    	username: $username,
                    	name: $name,
                    	surname: $surname,
                    	age: $age,
                    	password: $password,
                    	createdAt: $createdAt,
                    	updatedAt: $updatedAt })
                    RETURN user""", asParams(user));
            return fromRecord(result.single());
        };
    }

    private TransactionWork<Optional<User>> userRetrieval(int userId) {
        return tx -> {
            Result result = tx.run("""               
                    MATCH (user:User) WHERE ID(user) = $id
                    RETURN user""", Map.of("id", userId));

            if (!result.hasNext()) {
                return Optional.empty();
            }
            return Optional.of(fromRecord(result.single()));
        };
    }

    private TransactionWork<Boolean> userPasswordUpdate(int userId, PasswordUpdate passwordUpdate) {
        return tx -> {
            Result result = tx.run("""               
                    MATCH (user:User)
                    WHERE ID(user) = $id AND user.password = $old
                    SET user.password = $new
                    RETURN COUNT(user) = 1 AS successfulUpdate
            """, Map.of("id", userId, "old", passwordUpdate.getPassword(), "new", passwordUpdate.getNewPassword()));

            return result.single().get("successfulUpdate").asBoolean();
        };
    }

    private Map<String, Object> asParams(User user) {
        ZonedDateTime now = ZonedDateTime.now(ZoneId.of("UTC"));
        return Map.of(
                "username", user.getUsername(),
                "name", user.getName(),
                "surname", user.getSurname(),
                "age", user.getAge(),
                "password", user.getPassword(),
                "createdAt", now,
                "updatedAt", now
        );
    }

    private User fromRecord(Record record) {
        Node userNode = record.get("user").asNode();
        return new User(
                userNode.id(),
                userNode.get("username").asString(),
                userNode.get("name").asString(),
                userNode.get("age").asInt(),
                userNode.get("surname").asString(),
                userNode.get("password").asString(),
                userNode.get("createdAt").asZonedDateTime(),
                userNode.get("updatedAt").asZonedDateTime()
        );
    }

}
