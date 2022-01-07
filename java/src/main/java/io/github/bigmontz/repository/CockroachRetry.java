package io.github.bigmontz.repository;

import io.github.bigmontz.JdbcFunction;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.SQLException;

public class CockroachRetry {

    private static final int MAX_RETRY_COUNT = 3;
    private static final String RETRY_SQL_STATE = "40001";

    // adapted from https://www.cockroachlabs.com/docs/stable/build-a-java-app-with-cockroachdb.html
    public static <T> T retrySql(DataSource dataSource, JdbcFunction<Connection, T> work) {
        try (Connection connection = dataSource.getConnection()) {
            connection.setAutoCommit(false);
            int retryCount = 0;
            while (retryCount <= MAX_RETRY_COUNT) {
                if (retryCount == MAX_RETRY_COUNT) {
                    throw new RuntimeException(String.format("hit max of %s retries, aborting", MAX_RETRY_COUNT));
                }
                try {
                    T result = work.apply(connection);
                    connection.commit();
                    return result;
                } catch (SQLException e) {
                    if (!RETRY_SQL_STATE.equals(e.getSQLState())) {
                        throw e;
                    }
                    // Since this is a transaction retry error, we
                    // roll back the transaction and sleep a
                    // little before trying again.  Each time
                    // through the loop we sleep for a little
                    // longer than the last time
                    // (A.K.A. exponential backoff).
                    connection.rollback();
                    retryCount++;
                    int sleepMillis = (int) (Math.pow(2, retryCount) * 100); // except we removed the random addendum
                    try {
                        Thread.sleep(sleepMillis);
                    } catch (InterruptedException ignored) {
                        // Necessary to allow the Thread.sleep()
                        // above so the retry loop can continue.
                    }
                }
            }
        } catch (SQLException e) {
            throw new RuntimeException(e);
        }
        throw new RuntimeException("unreachable 🤷‍");
    }
}
