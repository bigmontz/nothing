package io.github.bigmontz.repository;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Timestamp;
import java.time.Instant;
import java.time.ZoneId;
import java.util.Optional;

public class UserPostgresRepository implements UserRepository<Long> {

    private static final ZoneId UTC = ZoneId.of("UTC");

    private final DataSource dataSource;

    public UserPostgresRepository(DataSource dataSource) {
        this.dataSource = dataSource;
    }

    @Override
    public Long parseId(String rawId) {
        return Long.parseLong(rawId);
    }

    @Override
    public String printId(Long value) {
        return value.toString();
    }

    @Override
    public User create(User user) {
        try (Connection connection = dataSource.getConnection();
             var preparedStatement = connection.prepareStatement("INSERT INTO users (username, name, age, surname, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING *")) {
            var now = now();
            preparedStatement.setString(1, user.getUsername());
            preparedStatement.setString(2, user.getName());
            preparedStatement.setInt(3, user.getAge());
            preparedStatement.setString(4, user.getSurname());
            preparedStatement.setString(5, user.getPassword());
            preparedStatement.setTimestamp(6, now);
            preparedStatement.setTimestamp(7, now);
            ResultSet resultSet = preparedStatement.executeQuery();
            return map(resultSet).orElseThrow(() -> new RuntimeException("could not create user"));
        } catch (SQLException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public Optional<User> findById(Long userId) {
        try (var connection = dataSource.getConnection();
             var statement = connection.prepareStatement("SELECT * FROM users WHERE id = ?")) {
            statement.setLong(1, userId);
            return map(statement.executeQuery());
        } catch (SQLException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public boolean updatePassword(Long userId, PasswordUpdate passwordUpdate) {
        // JDBC => no tx function / no retry 😢
        try (var connection = dataSource.getConnection()) {
            connection.setAutoCommit(false);
            return doUpdatePassword(connection, userId, passwordUpdate);
        } catch (SQLException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public void close() {

    }

    boolean doUpdatePassword(Connection connection, Long userId, PasswordUpdate passwordUpdate) throws SQLException {
        try (var statement = connection.prepareStatement("UPDATE users SET password = ?, updated_at = ? WHERE id = ? AND password = ?")) {
            statement.setString(1, passwordUpdate.getNewPassword());
            statement.setTimestamp(2, now());
            statement.setLong(3, userId);
            statement.setString(4, passwordUpdate.getPassword());
            if (statement.executeUpdate() == 1) {
                connection.commit();
                return true;
            }
            connection.rollback(); // covers the impossible case of matching more than 1 row (here be 🐉)
            return false;
        } catch (SQLException e) {
            connection.rollback();
            throw new RuntimeException(e);
        }
    }

    private static boolean hasNext(ResultSet resultSet) {
        try {
            return resultSet.next();
        } catch (SQLException e) {
            throw new RuntimeException("could not create user");
        }
    }

    private Optional<User> map(ResultSet resultSet) {
        if (!hasNext(resultSet)) {
            return Optional.empty();
        }
        try {
            return Optional.of(new User(
                    resultSet.getLong("id"),
                    resultSet.getString("username"),
                    resultSet.getString("name"),
                    resultSet.getInt("age"),
                    resultSet.getString("surname"),
                    resultSet.getString("password"),
                    resultSet.getTimestamp("created_at").toInstant().atZone(UTC),
                    resultSet.getTimestamp("updated_at").toInstant().atZone(UTC)
            ));
        } catch (SQLException e) {
            throw new RuntimeException(e);
        }
    }

    private static Timestamp now() {
        return new Timestamp(Instant.now().toEpochMilli());
    }

}
