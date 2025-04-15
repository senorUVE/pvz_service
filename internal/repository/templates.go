package repository

const (
	getIdFromUsers = `SELECT id from users WHERE email = $1`

	createUser = `INSERT INTO users (id, email, password_salt, role) VALUES ($1, $2, $3, $4) RETURNING id`

	insertUser = `INSERT INTO users ( email, password_salt, role) VALUES ($1, $2, $3) RETURNING id`

	getUserByEmail = `SELECT id, email, password_salt, role FROM users WHERE email = $1`

	userExists = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	createPVZ = `INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3) RETURNING id`

	getPVZWithReceptions = `SELECT p.id, p.registration_date, p.city,
                                r.id, r.date_time, r.status,
                                pr.id, pr.date_time, pr.type
                             FROM pvz p
                             LEFT JOIN reception r ON p.id = r.pvz_id
                             LEFT JOIN product pr ON r.id = pr.reception_id
                             WHERE ($1::timestamp IS NULL OR r.date_time >= $1)
                             AND ($2::timestamp IS NULL OR r.date_time <= $2)
                             ORDER BY p.registration_date
                             LIMIT $3 OFFSET $4`

	createReception = `INSERT INTO reception (id, date_time, pvz_id, status) VALUES ($1, $2, $3, 'in_progress') RETURNING id, date_time, status`

	closeLastReception = `UPDATE reception 
                            SET status = 'close' 
                            WHERE id = (
								SELECT r.id
								FROM reception r
								WHERE r.pvz_id = $1
								ORDER BY r.date_time DESC
								LIMIT 1
							)
							RETURNING id, date_time, pvz_id, status`

	getActiveReception = `SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id = $1 AND status = 'in_progress' LIMIT 1`

	getProductFromReception = `SELECT id FROM reception WHERE pvz_id = $1 AND status = 'in_progress' FOR UPDATE`

	createProduct = `INSERT INTO product (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4) RETURNING id`

	deleteLastProduct = `DELETE FROM product WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1 RETURNING id`

	deleteProduct = `DELETE FROM product WHERE id = (SELECT id FROM product WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1)`

	deleteLastProductQuery = `WITH active_reception AS (SELECT id FROM reception WHERE pvz_id = $1 AND status = 'in_progress' LIMIT 1) DELETE FROM product WHERE id = (SELECT id FROM product WHERE reception_id = (SELECT id FROM active_reception) ORDER BY date_time DESC LIMIT 1)`
)
