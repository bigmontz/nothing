package io.github.bigmontz.repository;

import javax.sql.DataSource;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Timestamp;
import java.time.Instant;
import java.time.ZoneId;
import java.util.Optional;

// same as UserPostgresRepository except for updatePassword
public class UserCockroachRepository implements UserRepository<Long> {

    private static final ZoneId UTC = ZoneId.of("UTC");

    private final DataSource dataSource;
    private final UserPostgresRepository delegate;

    public UserCockroachRepository(DataSource dataSource) {
        this.dataSource = dataSource;
        this.delegate = new UserPostgresRepository(dataSource);
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
        return delegate.create(user);
    }

    @Override
    public Optional<User> findById(Long userId) {
        return delegate.findById(userId);
    }

    @Override
    public boolean updatePassword(Long userId, PasswordUpdate passwordUpdate) {
        return CockroachRetry.retrySql(dataSource,
                connection -> delegate.doUpdatePassword(connection, userId, passwordUpdate));
    }

    @Override
    public void close() {

    }

}
