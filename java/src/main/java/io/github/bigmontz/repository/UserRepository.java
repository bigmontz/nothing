package io.github.bigmontz.repository;

import java.io.Closeable;
import java.util.Optional;

public interface UserRepository<ID> extends Closeable {

    ID parseId(String rawId);

    String printId(ID id);

    User create(User user);

    Optional<User> findById(ID userId);

    boolean updatePassword(ID userId, PasswordUpdate passwordUpdate);
}
