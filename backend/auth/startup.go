package auth

import (
	"context"
	"log"
	"os"
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
		var totalUsers int
		db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalUsers)
		if totalUsers > 0 {
			log.Println("[BOOTSTRAP] FATAL: Users exist but no admin found. Refusing to start.")
			os.Exit(1)
		}
		log.Println("[BOOTSTRAP] No users yet. First registrant will become admin.")
	}
}
