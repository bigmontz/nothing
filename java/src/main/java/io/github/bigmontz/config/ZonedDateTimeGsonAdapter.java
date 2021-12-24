package io.github.bigmontz.config;

import com.google.gson.JsonDeserializationContext;
import com.google.gson.JsonDeserializer;
import com.google.gson.JsonElement;
import com.google.gson.JsonParseException;
import com.google.gson.JsonPrimitive;
import com.google.gson.JsonSerializationContext;
import com.google.gson.JsonSerializer;

import java.lang.reflect.Type;
import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;

// needed since Gson reflection on ZonedDateTime breaks java.time module boundaries
public class ZonedDateTimeGsonAdapter implements JsonSerializer<ZonedDateTime>, JsonDeserializer<ZonedDateTime> {

    @Override
    public ZonedDateTime deserialize(JsonElement jsonElement, Type type, JsonDeserializationContext jsonDeserializationContext) throws JsonParseException {
        String value = jsonElement.getAsString();
        return ZonedDateTime.parse(value);
    }

    @Override
    public JsonElement serialize(ZonedDateTime value, Type type, JsonSerializationContext jsonSerializationContext) {
        return new JsonPrimitive(value.format(DateTimeFormatter.ISO_OFFSET_DATE_TIME));
    }
}