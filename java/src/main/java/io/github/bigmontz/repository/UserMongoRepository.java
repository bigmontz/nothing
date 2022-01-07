package io.github.bigmontz.repository;

import com.mongodb.client.ClientSession;
import com.mongodb.client.MongoClient;
import com.mongodb.client.MongoCollection;
import com.mongodb.client.model.Filters;
import org.bson.BsonObjectId;
import org.bson.Document;
import org.bson.conversions.Bson;
import org.bson.types.ObjectId;

import java.time.Instant;
import java.time.ZoneId;
import java.util.Date;
import java.util.Optional;

import static com.mongodb.client.model.Filters.and;
import static com.mongodb.client.model.Updates.set;

public class UserMongoRepository implements UserRepository<BsonObjectId> {

    private static final ZoneId UTC = ZoneId.of("UTC");

    private final MongoClient mongoClient;

    public UserMongoRepository(MongoClient mongoClient) {
        this.mongoClient = mongoClient;
    }

    @Override
    public BsonObjectId parseId(String rawId) {
        return new BsonObjectId(new ObjectId(rawId));
    }

    @Override
    public String printId(BsonObjectId bsonObjectId) {
        return bsonObjectId.getValue().toHexString();
    }

    @Override
    public User create(User user) {
        var result = userCollection().insertOne(toDocument(user));
        return findById(result.getInsertedId().asObjectId())
                .orElseThrow(() -> new RuntimeException("user creation failed"));
    }

    @Override
    public Optional<User> findById(BsonObjectId userId) {
        var document = userCollection().find(byObjectId(userId)).first();
        return Optional.ofNullable(document).map(UserMongoRepository::fromDocument);
    }

    @Override
    public boolean updatePassword(BsonObjectId userId, PasswordUpdate passwordUpdate) {
        try (ClientSession clientSession = mongoClient.startSession()) {
            return clientSession.withTransaction(() -> {
                Document result = userCollection()
                        .findOneAndUpdate(
                                and(byObjectId(userId), byPassword(passwordUpdate.getPassword())),
                                set("password", passwordUpdate.getNewPassword())
                        );
                return result != null;
            });
        }
    }

    @Override
    public void close() {
        mongoClient.close();
    }

    private static Document toDocument(User user) {
        var now = Instant.now();
        var document = new Document();
        document.put("username", user.getUsername());
        document.put("name", user.getName());
        document.put("age", user.getAge());
        document.put("surname", user.getSurname());
        document.put("password", user.getPassword());
        document.put("created_at", now);
        document.put("updated_at", now);
        return document;
    }

    private static User fromDocument(Document document) {
        return new User(
                document.get("_id", ObjectId.class).toHexString(),
                document.get("username", String.class),
                document.get("name", String.class),
                document.get("age", Integer.class),
                document.get("surname", String.class),
                document.get("password", String.class),
                document.get("created_at", Date.class).toInstant().atZone(UTC),
                document.get("updated_at", Date.class).toInstant().atZone(UTC)
        );
    }

    private MongoCollection<Document> userCollection() {
        return mongoClient.getDatabase("admin").getCollection("users");
    }

    private static Bson byObjectId(BsonObjectId id) {
        return Filters.eq("_id", id);
    }

    private static Bson byPassword(String password) {
        return Filters.eq("password", password);
    }
}
