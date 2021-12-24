package io.github.bigmontz.repository;

import java.io.Closeable;
import java.util.Optional;

public interface UserRepository extends Closeable {

    User create(User user);

    Optional<User> findById(Object userId);

    boolean updatePassword(Object userId, PasswordUpdate passwordUpdate);
}
