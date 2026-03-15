package eloq

//
//func TestSelectBuilder_PostgreSQLStyle(t *testing.T) {
//	sql, _, err := commonBuilder.Select("id").
//		From("users").
//		PlaceholderFormat(eloquentgo.Dollar).
//		QuoteWith(eloquentgo.DoubleQuote).
//		WhereIn("status", []interface{}{"active", "pending"}).
//		ToSql()
//
//	assert.NoError(t, err)
//	assert.Equal(t, `SELECT "id" FROM "users" WHERE "status" IN ($1, $2)`, sql)
//}

//func TestUpdateBuilder(t *testing.T) {
//	now := time.Now()
//
//	sql, args, err := eloquentgo.Update("users").
//		Set("status", "banned").
//		Set("banned_at", now).
//		Set("score", nil).
//		Where("id", 42).
//		Where("last_login", "<", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)).
//		Limit(1).
//		ToSql()
//
//	assert.NoError(t, err)
//	assert.Equal(t, `UPDATE "users" SET "status" = ?, "banned_at" = ?, "score" = NULL WHERE "id" = ? AND "last_login" < ? LIMIT 1`, sql)
//	assert.Equal(t, []interface{}{"banned", now, 42, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}, args)
//}
//
//func TestUpdateBuilder_WithWhen(t *testing.T) {
//	shouldBan := false
//
//	sql, _, err := eloquentgo.Update("users").
//		Set("updated_at", time.Now()).
//		When(shouldBan, func(q *eloquentgo.commonBuilder) *eloquentgo.commonBuilder {
//			return q.Set("status", "banned")
//		}, func(q *eloquentgo.commonBuilder) *eloquentgo.commonBuilder {
//			return q.Set("status", "active")
//		}).
//		Where("id", 100).
//		ToSql()
//
//	assert.NoError(t, err)
//	assert.Contains(t, sql, `"status" = ?`)
//	// status должен быть "active", потому что shouldBan == false
//	assert.Contains(t, sql, `"status" = 'active'"`) // нет, плейсхолдер, но значение в args будет "active"
//	// лучше проверить через отдельный тест с захватом args
//}
//
//func TestDeleteBuilder(t *testing.T) {
//	sql, args, err := eloquentgo.Delete("sessions").
//		Where("user_id", 999).
//		Where("expires_at", "<", time.Now()).
//		OrderBy("created_at", "asc").
//		Limit(500).
//		ToSql()
//
//	assert.NoError(t, err)
//	assert.Equal(t, `DELETE FROM "sessions" WHERE "user_id" = ? AND "expires_at" < ? ORDER BY "created_at" ASC LIMIT 500`, sql)
//	assert.Len(t, args, 2)
//}
//
//func TestInsertBuilder_SingleRow(t *testing.T) {
//	sql, args, err := eloquentgo.InsertInto("users").
//		Columns("name", "email", "password_hash").
//		Values("Alice", "alice@example.com", "hashedpass123").
//		ToSql()
//
//	assert.NoError(t, err)
//	assert.Equal(t, `INSERT INTO "users" ("name", "email", "password_hash") VALUES (?, ?, ?)`, sql)
//	assert.Equal(t, []interface{}{"Alice", "alice@example.com", "hashedpass123"}, args)
//}
//
//func TestInsertBuilder_MultipleRows(t *testing.T) {
//	sql, args, err := eloquentgo.InsertInto("posts").
//		Columns("user_id", "title", "body").
//		Values(1, "First Post", "Hello world!").
//		Values(2, "Second", "Go is awesome").
//		Values(3, nil, "Draft").
//		ToSql()
//
//	assert.NoError(t, err)
//	assert.Equal(t, `INSERT INTO "posts" ("user_id", "title", "body") VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)`, sql)
//	assert.Equal(t, []interface{}{1, "First Post", "Hello world!", 2, "Second", "Go is awesome", 3, nil, "Draft"}, args)
//}
//
//func TestInsertBuilder_ValueMap(t *testing.T) {
//	sql, args, err := eloquentgo.InsertInto("profiles").
//		ValueMap(map[string]interface{}{
//			"user_id":   42,
//			"bio":       "Golang enthusiast",
//			"avatar":    nil,
//			"is_public": true,
//		}).
//		ValueMap(map[string]interface{}{
//			"user_id":   43,
//			"bio":       "New user",
//			"is_public": false,
//		}).
//		ToSql()
//
//	assert.NoError(t, err)
//	assert.Equal(t, `INSERT INTO "profiles" ("user_id", "bio", "avatar", "is_public") VALUES (?, ?, ?, ?), (?, ?, ?, ?)`, sql)
//	assert.Equal(t, []interface{}{42, "Golang enthusiast", nil, true, 43, "New user", nil, false}, args)
//}
