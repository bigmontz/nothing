package io.github.bigmontz.config;

import java.util.function.Function;
import java.util.function.Supplier;

import static java.util.function.Function.identity;

public class Env {

    public static String getOrDefault(String name, String defaultValue) {
        return getOrDefault(name, defaultValue, identity());
    }

    public static <T> T getOrDefault(String name, T defaultValue, Function<String, T> fn) {
        String value = System.getenv(name);
        if (value == null) {
            return defaultValue;
        }
        return fn.apply(value);
    }

    public static <T extends RuntimeException> String getOrThrow(String name, Supplier<T> supplier) {
        String value = System.getenv(name);
        if (value == null) {
            throw supplier.get();
        }
        return value;
    }
}
