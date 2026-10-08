package database

import "fmt"

// SeedCafes inserts sample cafes when the cafes table is empty.
func SeedCafes() error {
	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM cafes`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	cafes := []struct{ name, address, description string }{
		{"Bean There", "12 Market St", "Single-origin pour-overs and a quiet back room."},
		{"Morning Crumb", "48 Elm Ave", "Fresh pastries baked before sunrise."},
		{"Corner Roast", "7 Harbor Rd", "Espresso bar with a good selection of cold brew."},
	}

	for _, c := range cafes {
		if _, err := DB.Exec(
			`INSERT INTO cafes (name, address, description) VALUES ($1, $2, $3)`,
			c.name, c.address, c.description,
		); err != nil {
			return fmt.Errorf("seed cafe %q: %w", c.name, err)
		}
	}
	return nil
}
