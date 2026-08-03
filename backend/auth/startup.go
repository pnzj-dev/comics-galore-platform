package auth

import (
	"context"
	"log"
)

func init() {
	checkAdminExists()
}

func checkAdminExists() {
	ctx := context.Background()

	var count int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	if err != nil {
		log.Printf("[BOOTSTRAP] could not check for admin users: %v", err)
		return
	}

	if count == 0 {
		log.Println("[BOOTSTRAP] FATAL: No admin user exists in the database.")
		log.Println("[BOOTSTRAP] The application cannot start without at least one admin.")
		log.Println("[BOOTSTRAP] The first user to register will automatically be assigned the 'admin' role.")
	}
}
