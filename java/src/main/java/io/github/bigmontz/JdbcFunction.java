package io.github.bigmontz;

import java.sql.SQLException;

@FunctionalInterface
public interface JdbcFunction<T, R> {
    /**
     * Applies this function to the given argument.
     *
     * @param t the function argument
     * @return the function result
     */
    R apply(T t) throws SQLException;
}
