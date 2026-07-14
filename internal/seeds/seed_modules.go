package seeds

import (
	"encoding/json"
	"fmt"
	"log"

	"pleco-api/internal/modules/event"
	"pleco-api/internal/modules/guide"
	"pleco-api/internal/modules/hotel"
	"pleco-api/internal/modules/partner"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/rental"
	"pleco-api/internal/modules/restaurant"
	"pleco-api/internal/modules/review"
	"pleco-api/internal/modules/souvenir"
	"pleco-api/internal/modules/story"

	"gorm.io/gorm"
)

func SeedModules(db *gorm.DB) {
	mustHaveDB(db)

	seedEvents(db)
	seedHotels(db)
	seedRestaurants(db)
	seedPartners(db)
	seedGuides(db)
	seedSouvenirs(db)
	seedRentals(db)
	seedReviews(db)
	seedStories(db)
	seedPromotions(db)
}

func hja(v interface{}) hotel.JSONArr {
	b, _ := json.Marshal(v)
	var arr hotel.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func rja(v interface{}) restaurant.JSONArr {
	b, _ := json.Marshal(v)
	var arr restaurant.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func gja(v interface{}) guide.JSONArr {
	b, _ := json.Marshal(v)
	var arr guide.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func sja(v interface{}) souvenir.JSONArr {
	b, _ := json.Marshal(v)
	var arr souvenir.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func rntja(v interface{}) rental.JSONArr {
	b, _ := json.Marshal(v)
	var arr rental.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func rvja(v interface{}) review.JSONArr {
	b, _ := json.Marshal(v)
	var arr review.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func stja(v interface{}) story.JSONArr {
	b, _ := json.Marshal(v)
	var arr story.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func eja(v interface{}) event.JSONArr {
	b, _ := json.Marshal(v)
	var arr event.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func seedEvents(db *gorm.DB) {
	// Use upsert logic so new festival events can be added even if table is non-empty
	allEvents := []event.Event{
		{
			ExternalID:   "evt-1",
			Title:        "Yogyakarta Jazz Festival",
			Description:  "Annual jazz music festival featuring local and international artists at the stunning Prambanan temple complex.",
			Location:     "Prambanan Temple, Sleman",
			StartDate:    "2026-07-20",
			EndDate:      "2026-07-22",
			ImageURL:     "https://images.unsplash.com/photo-1459749411175-04bf5292ceea?auto=format&fit=crop&w=1200&q=80",
			Category:     "music",
			Status:       "upgoing",
			Latitude:     -7.7520,
			Longitude:    110.4914,
			MaxAttendees: 5000,
			TicketPrice:  "IDR 350,000",
			Organizer:    "Jogja Jazz Foundation",
			Highlights:   eja([]string{"Live International Acts", "Prambanan Night Stage", "Traditional Jazz Fusion"}),
		},
		{
			ExternalID:   "evt-2",
			Title:        "Bakpia Trail Festival",
			Description:  "Celebrate Yogyakarta's iconic Bakpia pastry with cooking workshops, tastings, and cultural performances.",
			Location:     "Pathuk, Yogyakarta",
			StartDate:    "2026-08-15",
			EndDate:      "2026-08-16",
			ImageURL:     "https://images.unsplash.com/photo-1504674900247-0877df9cc836?auto=format&fit=crop&w=1200&q=80",
			Category:     "food",
			Status:       "upgoing",
			Latitude:     -7.7830,
			Longitude:    110.3700,
			MaxAttendees: 2000,
			TicketPrice:  "IDR 100,000",
			Organizer:    "Yogyakarta Culinary Board",
			Highlights:   eja([]string{"Bakpia Cooking Workshop", "Street Food Market", "Cultural Performances"}),
		},
		{
			ExternalID:   "evt-3",
			Title:        "Saman Dance Festival",
			Description:  "Annual showcase of the traditional Saman dance from Gayo, performed by troupes across Indonesia.",
			Location:     "Kraton Complex, Yogyakarta",
			StartDate:    "2026-09-01",
			EndDate:      "2026-09-03",
			ImageURL:     "https://images.unsplash.com/photo-1545156521-77bd85671d30?auto=format&fit=crop&w=1200&q=80",
			Category:     "cultural",
			Status:       "upgoing",
			Latitude:     -7.8054,
			Longitude:    110.3642,
			MaxAttendees: 3000,
			TicketPrice:  "IDR 50,000",
			Organizer:    "Yogyakarta Cultural Office",
			Highlights:   eja([]string{"Traditional Saman Dance", "Multi-Troupe Performance", "Gamelan Orchestra"}),
		},
		// --- FESTIVALS from frontend static data (merged) ---
		{
			ExternalID:   "f-sekaten",
			Title:        "Sekaten Festival",
			Description:  "A week-long royal and spiritual festival celebrating the birth of Prophet Muhammad. The sacred royal gamelans are carried from the palace to the Grand Mosque.",
			Location:     "Yogyakarta",
			StartDate:    "2025-05-07",
			EndDate:      "2025-05-15",
			ImageURL:     "https://images.unsplash.com/photo-1514933651103-005eec06c04b?q=80&w=1200",
			Category:     "culture",
			Status:       "upgoing",
			Latitude:     -7.8054,
			Longitude:    110.3642,
			MaxAttendees: 10000,
			TicketPrice:  "Free",
			Organizer:    "Keraton Yogyakarta",
			Highlights:   eja([]string{"Royal Gamelan Processions", "Traditional Javanese Night Market"}),
		},
		{
			ExternalID:   "f-fky",
			Title:        "Jogja Art Festival",
			Description:  "The ultimate showcase of contemporary and traditional art, bringing Yogyakarta's streets to life with street carnivals and puppet plays.",
			Location:     "Yogyakarta",
			StartDate:    "2025-06-23",
			EndDate:      "2025-06-30",
			ImageURL:     "https://images.unsplash.com/photo-1501281668745-f7f57925c3b4?q=80&w=1200",
			Category:     "art",
			Status:       "upgoing",
			Latitude:     -7.7928,
			Longitude:    110.3658,
			MaxAttendees: 8000,
			TicketPrice:  "Free",
			Organizer:    "Festival Kesenian Yogyakarta",
			Highlights:   eja([]string{"Street Art Carnivals", "Wayang Kulit Puppetry Night"}),
		},
		{
			ExternalID:   "f-grebeg",
			Title:        "Grebeg Maulud",
			Description:  "The spectacular peak of royal gratitude. The Sultan of Yogyakarta paraded colossal mountain-shaped offerings made of harvest crops.",
			Location:     "Yogyakarta",
			StartDate:    "2025-10-05",
			EndDate:      "2025-10-05",
			ImageURL:     "https://images.unsplash.com/photo-1533174072545-7a4b6ad7a6c3?q=80&w=1200",
			Category:     "culture",
			Status:       "upgoing",
			Latitude:     -7.8054,
			Longitude:    110.3642,
			MaxAttendees: 15000,
			TicketPrice:  "Free",
			Organizer:    "Keraton Yogyakarta",
			Highlights:   eja([]string{"Majestic Gunungan Mountains", "Ten Royal Regiments"}),
		},
		{
			ExternalID:   "f-wonosari",
			Title:        "Wonosari Night Carnival",
			Description:  "A vibrant and glowing street parade in Gunungkidul featuring traditional dances, music, and colorful illuminated cultural floats.",
			Location:     "Gunungkidul",
			StartDate:    "2025-07-12",
			EndDate:      "2025-07-12",
			ImageURL:     "https://images.unsplash.com/photo-1516450360452-9312f5e86fc7?q=80&w=1200",
			Category:     "carnival",
			Status:       "upgoing",
			Latitude:     -7.9800,
			Longitude:    110.5900,
			MaxAttendees: 5000,
			TicketPrice:  "Free",
			Organizer:    "Pemerintah Gunungkidul",
			Highlights:   eja([]string{"Glowing Parades", "Night Carnival"}),
		},
	}

	inserted := 0
	updated := 0
	for _, e := range allEvents {
		var existing event.Event
		err := db.Where("external_id = ?", e.ExternalID).First(&existing).Error
		if err != nil {
			if dbErr := db.Create(&e).Error; dbErr != nil {
				log.Printf("Failed to create event %s: %v", e.ExternalID, dbErr)
			} else {
				inserted++
			}
		} else {
			e.ID = existing.ID
			e.CreatedAt = existing.CreatedAt
			if dbErr := db.Save(&e).Error; dbErr != nil {
				log.Printf("Failed to update event %s: %v", e.ExternalID, dbErr)
			} else {
				updated++
			}
		}
	}
	fmt.Printf("Events seeding done: %d inserted, %d updated\n", inserted, updated)
}

func seedHotels(db *gorm.DB) {
	var count int64
	db.Model(&hotel.Hotel{}).Count(&count)
	if count > 0 {
		fmt.Println("Hotels already seeded, skipping")
		return
	}

	hotels := []hotel.Hotel{
		{
			ExternalID:    "htl-1",
			Name:          "The Phoenix Hotel Yogyakarta",
			Description:   "A luxurious colonial heritage hotel with royal Javanese spa and grand courtyard pool.",
			Location:      "Yogyakarta City",
			Address:       "Jl. Jend. Sudirman No.9, Yogyakarta",
			Stars:         5,
			PricePerNight: "IDR 1,500,000",
			Images:        hja([]string{"https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600"}),
			Amenities:     hja([]string{"Pool", "Spa", "Restaurant", "WiFi", "Parking"}),
			Phone:         "+62 274 123456",
			Email:         "info@phoenixjogja.com",
			Website:       "https://phoenixjogja.com",
			Rating:        4.9,
			ReviewCount:   1240,
			Latitude:      -7.7895,
			Longitude:     110.3642,
		},
		{
			ExternalID:    "htl-2",
			Name:          "Royal Ambarrukmo Hotel",
			Description:   "Heritage luxury hotel once used by visiting royalty, featuring Javanese architecture.",
			Location:      "Yogyakarta City",
			Address:       "Jl. Babarsari No.53, Yogyakarta",
			Stars:         5,
			PricePerNight: "IDR 1,200,000",
			Images:        hja([]string{"https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=600"}),
			Amenities:     hja([]string{"Pool", "Spa", "Restaurant", "Bar", "WiFi"}),
			Phone:         "+62 274 234567",
			Email:         "reservations@royalambarrukmo.com",
			Website:       "https://royalambarrukmo.com",
			Rating:        4.8,
			ReviewCount:   980,
			Latitude:      -7.7820,
			Longitude:     110.3950,
		},
		{
			ExternalID:    "htl-3",
			Name:          "Queen of the South Resort",
			Description:   "Clifftop resort with infinity pool overlooking the Indian Ocean at Parangtritis.",
			Location:      "Bantul, Yogyakarta",
			Address:       "Parangrejo, Purwosari, Bantul",
			Stars:         4,
			PricePerNight: "IDR 1,400,000",
			Images:        hja([]string{"https://images.unsplash.com/photo-1571896349842-33c89424de2d?q=80&w=600"}),
			Amenities:     hja([]string{"Infinity Pool", "Restaurant", "Ocean View", "WiFi"}),
			Phone:         "+62 274 345678",
			Email:         "info@queenofthesouth.com",
			Website:       "https://queenofthesouth.com",
			Rating:        4.8,
			ReviewCount:   756,
			Latitude:      -8.0253,
			Longitude:     110.3298,
		},
	}
	if err := db.CreateInBatches(hotels, 10).Error; err != nil {
		log.Printf("Failed to seed hotels: %v", err)
	}
	fmt.Printf("Hotels seeding done (%d records)\n", len(hotels))
}

func seedRestaurants(db *gorm.DB) {
	var count int64
	db.Model(&restaurant.Restaurant{}).Count(&count)
	if count > 0 {
		fmt.Println("Restaurants already seeded, skipping")
		return
	}

	restaurants := []restaurant.Restaurant{
		{
			ExternalID:   "rst-1",
			Name:         "Gudeg Yu Djum",
			Description:  "The most legendary gudeg restaurant in Yogyakarta since 1950.",
			Location:     "Yogyakarta City",
			Address:      "Jl. Wijaya Kusuma No.20, Yogyakarta",
			CuisineType:  "Javanese",
			PriceRange:   "$",
			Images:       rja([]string{"https://images.unsplash.com/photo-1504674900247-0877df9cc836?q=80&w=600"}),
			OpeningHours: "08:00 AM - 10:00 PM",
			Phone:        "+62 274 456789",
			Rating:       4.8,
			ReviewCount:  3200,
			Latitude:     -7.7950,
			Longitude:    110.3690,
		},
		{
			ExternalID:   "rst-2",
			Name:         "Abhayagiri Restaurant",
			Description:  "Hilltop fine dining with panoramic sunset views of Prambanan and Mount Merapi.",
			Location:     "Sleman, Yogyakarta",
			Address:      "Samberwatu, Sambirejo, Prambanan",
			CuisineType:  "Indonesian Fine Dining",
			PriceRange:   "$$$",
			Images:       rja([]string{"https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?q=80&w=600"}),
			OpeningHours: "11:00 AM - 11:00 PM",
			Phone:        "+62 274 567890",
			Rating:       4.8,
			ReviewCount:  1560,
			Latitude:     -7.7600,
			Longitude:    110.4850,
		},
		{
			ExternalID:   "rst-3",
			Name:         "Timang Lobster Warung",
			Description:  "Fresh grilled lobster caught daily by local fishermen at Timang Beach.",
			Location:     "Gunungkidul, Yogyakarta",
			Address:      "Timang Beach, Tepus, Gunungkidul",
			CuisineType:  "Seafood",
			PriceRange:   "$$",
			Images:       rja([]string{"https://images.unsplash.com/photo-1504674900247-0877df9cc836?q=80&w=600"}),
			OpeningHours: "08:00 AM - 05:00 PM",
			Phone:        "+62 274 678901",
			Rating:       4.7,
			ReviewCount:  890,
			Latitude:     -8.1350,
			Longitude:    110.6580,
		},
	}
	if err := db.CreateInBatches(restaurants, 10).Error; err != nil {
		log.Printf("Failed to seed restaurants: %v", err)
	}
	fmt.Printf("Restaurants seeding done (%d records)\n", len(restaurants))
}

func seedPartners(db *gorm.DB) {
	var count int64
	db.Model(&partner.Partner{}).Count(&count)
	if count > 0 {
		fmt.Println("Partners already seeded, skipping")
		return
	}

	partners := []partner.Partner{
		{
			ExternalID:  "ptr-1",
			Name:        "The Phoenix Hotel Yogyakarta",
			Description: "A luxurious colonial heritage hotel with royal Javanese spa and grand courtyard pool.",
			Category:    "hotel",
			Location:    "Yogyakarta City",
			Address:     "Jl. Jend. Sudirman No.9, Yogyakarta",
			Image:       "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600",
			Rating:      4.9,
			Price:       "IDR 1,500,000 / night",
			Distance:    "0 km",
			Phone:       "+62 274 123456",
			Website:     "https://phoenixjogja.com",
			Latitude:    -7.7895,
			Longitude:   110.3642,
		},
		{
			ExternalID:  "ptr-2",
			Name:        "Gunungkidul Cave Guides Guild",
			Description: "Highly trained SRT safety experts with decades of cave experience.",
			Category:    "guide",
			Location:    "Gunungkidul, Yogyakarta",
			Address:     "Pacarejo, Semanu, Gunungkidul",
			Image:       "https://images.unsplash.com/photo-1544717297-fa95b6ee9643?q=80&w=600",
			Rating:      4.9,
			Price:       "IDR 500,000 / trip",
			Distance:    "35 km",
			Phone:       "+62 274 789012",
			Website:     "",
			Latitude:    -8.0287,
			Longitude:   110.6384,
		},
		{
			ExternalID:  "ptr-3",
			Name:        "Borobudur Sunset Jeep Tour",
			Description: "Off-road jeep adventure to the best sunset viewpoints around Borobudur.",
			Category:    "transport",
			Location:    "Magelang, Central Java",
			Address:     "Jl. Badrawati, Borobudur, Magelang",
			Image:       "https://images.unsplash.com/photo-1544717297-fa95b6ee9643?q=80&w=600",
			Rating:      4.7,
			Price:       "IDR 350,000 / jeep",
			Distance:    "42 km",
			Phone:       "+62 293 890123",
			Website:     "https://borobudurjeep.com",
			Latitude:    -7.6079,
			Longitude:   110.2038,
		},
	}
	if err := db.CreateInBatches(partners, 10).Error; err != nil {
		log.Printf("Failed to seed partners: %v", err)
	}
	fmt.Printf("Partners seeding done (%d records)\n", len(partners))
}

func seedGuides(db *gorm.DB) {
	var count int64
	db.Model(&guide.Guide{}).Count(&count)
	if count > 0 {
		fmt.Println("Guides already seeded, skipping")
		return
	}

	guides := []guide.Guide{
		{
			ExternalID:     "gde-1",
			Name:           "Pak Surya Wijaya",
			Bio:            "15-year veteran guide specializing in Yogyakarta's royal heritage and temple complexes.",
			Specialization: "Heritage & Culture",
			Phone:          "+62 812 3456 7890",
			Email:          "surya@jogjaguide.com",
			Rating:         4.9,
			ReviewCount:    420,
			Languages:      gja([]string{"English", "Indonesian", "Japanese"}),
			PricePerDay:    "IDR 500,000",
			Avatar:         "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150",
		},
		{
			ExternalID:     "gde-2",
			Name:           "Mbak Dewi Lestari",
			Bio:            "Adventure guide certified in SRT caving and rock climbing. Expert in Gunungkidul caves.",
			Specialization: "Adventure & Caving",
			Phone:          "+62 813 4567 8901",
			Email:          "dewi@jogjaguide.com",
			Rating:         4.8,
			ReviewCount:    280,
			Languages:      gja([]string{"English", "Indonesian"}),
			PricePerDay:    "IDR 600,000",
			Avatar:         "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150",
		},
		{
			ExternalID:     "gde-3",
			Name:           "Mas Adi Pratama",
			Bio:            "Culinary tour specialist. Shows travelers the authentic hidden food gems of Yogyakarta.",
			Specialization: "Culinary Tours",
			Phone:          "+62 815 5678 9012",
			Email:          "adi@jogjaguide.com",
			Rating:         4.7,
			ReviewCount:    195,
			Languages:      gja([]string{"English", "Indonesian", "Mandarin"}),
			PricePerDay:    "IDR 450,000",
			Avatar:         "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=150",
		},
	}
	if err := db.CreateInBatches(guides, 10).Error; err != nil {
		log.Printf("Failed to seed guides: %v", err)
	}
	fmt.Printf("Guides seeding done (%d records)\n", len(guides))
}

func seedSouvenirs(db *gorm.DB) {
	var count int64
	db.Model(&souvenir.Souvenir{}).Count(&count)
	if count > 0 {
		fmt.Println("Souvenirs already seeded, skipping")
		return
	}

	souvenirs := []souvenir.Souvenir{
		{
			ExternalID:   "sou-1",
			Name:         "Mirota Batik",
			Description:  "The most famous batik store in Malioboro with thousands of handmade batik pieces.",
			Location:     "Yogyakarta City",
			Address:      "Jl. Malioboro No.28, Yogyakarta",
			Images:       sja([]string{"https://images.unsplash.com/photo-1581456495146-65a71b2c8e52?auto=format&fit=crop&w=1200&q=80"}),
			ProductTypes: sja([]string{"Batik Fabric", "Batik Clothing", "Batik Accessories"}),
			PriceRange:   "$$",
			Phone:        "+62 274 111222",
			Rating:       4.7,
			Latitude:     -7.7928,
			Longitude:    110.3658,
		},
		{
			ExternalID:   "sou-2",
			Name:         "Bakpia Pathok 75",
			Description:  "One of the original Bakpia Pathok shops producing the iconic Yogyakarta pastry since 1948.",
			Location:     "Yogyakarta City",
			Address:      "Jl. Podobudow No.75, Yogyakarta",
			Images:       sja([]string{"https://images.unsplash.com/photo-1504674900247-0877df9cc836?auto=format&fit=crop&w=1200&q=80"}),
			ProductTypes: sja([]string{"Bakpia", "Traditional Pastries", "Souvenir Boxes"}),
			PriceRange:   "$",
			Phone:        "+62 274 222333",
			Rating:       4.6,
			Latitude:     -7.7830,
			Longitude:    110.3700,
		},
		{
			ExternalID:   "sou-3",
			Name:         "Kasongan Pottery Village",
			Description:  "Traditional pottery village where artisans handcraft terracotta and ceramic art pieces.",
			Location:     "Bantul, Yogyakarta",
			Address:      "Jl. Kadisowo, Kasongan, Bantul",
			Images:       sja([]string{"https://images.unsplash.com/photo-1565193566173-7a0ee3dbe261?auto=format&fit=crop&w=1200&q=80"}),
			ProductTypes: sja([]string{"Pottery", "Ceramics", "Terracotta Art", "Vases"}),
			PriceRange:   "$-$$",
			Phone:        "+62 274 333444",
			Rating:       4.5,
			Latitude:     -7.8100,
			Longitude:    110.3500,
		},
	}
	if err := db.CreateInBatches(souvenirs, 10).Error; err != nil {
		log.Printf("Failed to seed souvenirs: %v", err)
	}
	fmt.Printf("Souvenirs seeding done (%d records)\n", len(souvenirs))
}

func seedRentals(db *gorm.DB) {
	var count int64
	db.Model(&rental.Rental{}).Count(&count)
	if count > 0 {
		fmt.Println("Rentals already seeded, skipping")
		return
	}

	rentals := []rental.Rental{
		{
			ExternalID:   "rnt-1",
			Name:         "Jogja Motor Rent",
			Description:  "Wide selection of motorbikes and cars for rent with free delivery to hotels.",
			Location:     "Yogyakarta City",
			Address:      "Jl. Sisingamangaraja No.15, Yogyakarta",
			VehicleTypes: rntja([]string{"Scooter", "Motorbike", "City Car", "SUV"}),
			PricePerDay:  "IDR 75,000",
			Images:       rntja([]string{"https://images.unsplash.com/photo-1558618666-fcd25c85f82e?auto=format&fit=crop&w=1200&q=80"}),
			Phone:        "+62 812 111222",
			Rating:       4.6,
			Latitude:     -7.7950,
			Longitude:    110.3690,
		},
		{
			ExternalID:   "rnt-2",
			Name:         "Sewa Jeep Merapi",
			Description:  "Specialized jeep rental for Mount Merapi lava tours with experienced local drivers.",
			Location:     "Sleman, Yogyakarta",
			Address:      "Jl. Kaliurang Km 16, Pakem, Sleman",
			VehicleTypes: rntja([]string{"Willys Jeep", "Hardtop Jeep", "4x4 SUV"}),
			PricePerDay:  "IDR 350,000",
			Images:       rntja([]string{"https://images.unsplash.com/photo-1544717297-fa95b6ee9643?auto=format&fit=crop&w=1200&q=80"}),
			Phone:        "+62 813 333444",
			Rating:       4.8,
			Latitude:     -7.6600,
			Longitude:    110.4200,
		},
		{
			ExternalID:   "rnt-3",
			Name:         "Andong Malioboro",
			Description:  "Traditional horse-drawn carriage rental for sightseeing tours around the Sultan's Palace.",
			Location:     "Yogyakarta City",
			Address:      "Jl. Malioboro, Yogyakarta",
			VehicleTypes: rntja([]string{"Andong (Horse Carriage)", "Bicycle Rickshaw"}),
			PricePerDay:  "IDR 200,000",
			Images:       rntja([]string{"https://images.unsplash.com/photo-1543874768-af0b9c4090d5?auto=format&fit=crop&w=1200&q=80"}),
			Phone:        "+62 815 555666",
			Rating:       4.5,
			Latitude:     -7.7928,
			Longitude:    110.3658,
		},
	}
	if err := db.CreateInBatches(rentals, 10).Error; err != nil {
		log.Printf("Failed to seed rentals: %v", err)
	}
	fmt.Printf("Rentals seeding done (%d records)\n", len(rentals))
}

func seedReviews(db *gorm.DB) {
	var count int64
	db.Model(&review.Review{}).Count(&count)
	if count > 0 {
		fmt.Println("Reviews already seeded, skipping")
		return
	}

	reviews := []review.Review{
		{
			ExternalID:    "rev-1",
			UserID:        "usr-admin",
			DestinationID: "prambanan",
			UserName:      "Sophia Laurent",
			Rating:        5,
			Comment:       "An absolutely magical experience. The architecture is breathtaking and the Ramayana Ballet was unforgettable.",
			Images:        rvja([]string{}),
			Status:        "approved",
		},
		{
			ExternalID:    "rev-2",
			UserID:        "usr-admin",
			DestinationID: "malioboro",
			UserName:      "Yuki Tanaka",
			Rating:        4,
			Comment:       "The energy of Malioboro at night is electric. Street food, music, and the friendliest people.",
			Images:        rvja([]string{}),
			Status:        "approved",
		},
		{
			ExternalID:    "rev-3",
			UserID:        "usr-admin",
			DestinationID: "merapi",
			UserName:      "Budi Santoso",
			Rating:        5,
			Comment:       "The sunrise jeep tour was the highlight of our Jogja trip. Absolutely thrilling!",
			Images:        rvja([]string{}),
			Status:        "approved",
		},
	}
	if err := db.CreateInBatches(reviews, 10).Error; err != nil {
		log.Printf("Failed to seed reviews: %v", err)
	}
	fmt.Printf("Reviews seeding done (%d records)\n", len(reviews))
}

func seedStories(db *gorm.DB) {
	var count int64
	db.Model(&story.Story{}).Count(&count)
	if count > 0 {
		fmt.Println("Stories already seeded, skipping")
		return
	}

	stories := []story.Story{
		{
			ExternalID:    "sty-1",
			UserID:        "usr-admin",
			Title:         "My First Sunrise at Prambanan",
			Content:       "Waking up at 4 AM was worth it. Watching the sun rise behind the ancient Hindu spires of Prambanan Temple was one of the most magical moments of my life. The golden light painted the stone carvings in ways I never imagined possible.",
			Images:        stja([]string{"https://images.unsplash.com/photo-1578469550956-0e16b69c6a3d?auto=format&fit=crop&w=1200&q=80"}),
			DestinationIDs: stja([]string{"prambanan"}),
			Likes:         234,
			Status:        "published",
		},
		{
			ExternalID:    "sty-2",
			UserID:        "usr-admin",
			Title:         "Caving Into the Light of Heaven",
			Content:       "Rappelling 60 meters into the darkness of Goa Jomblang was terrifying and exhilarating. But when that single beam of sunlight pierced through the cave ceiling at noon, illuminating the underground forest below, I forgot to breathe. It truly felt like heaven.",
			Images:        stja([]string{"https://images.unsplash.com/photo-1628047563315-d1e8b8d222b9?auto=format&fit=crop&w=1200&q=80"}),
			DestinationIDs: stja([]string{"goajomblang"}),
			Likes:         189,
			Status:        "published",
		},
		{
			ExternalID:    "sty-3",
			UserID:        "usr-admin",
			Title:         "Malioboro After Dark",
			Content:       "As the sun sets, Malioboro transforms. The hum of motorbikes gives way to the clatter of horse-drawn carriages, the sizzle of street food vendors, and the melody of buskers playing traditional Javanese music. This street is truly the soul of Yogyakarta.",
			Images:        stja([]string{"https://images.unsplash.com/photo-1543874768-af0b9c4090d5?auto=format&fit=crop&w=1200&q=80"}),
			DestinationIDs: stja([]string{"malioboro"}),
			Likes:         156,
			Status:        "published",
		},
	}
	if err := db.CreateInBatches(stories, 10).Error; err != nil {
		log.Printf("Failed to seed stories: %v", err)
	}
	fmt.Printf("Stories seeding done (%d records)\n", len(stories))
}

func seedPromotions(db *gorm.DB) {
	var count int64
	db.Model(&promotion.Promotion{}).Count(&count)
	if count > 0 {
		fmt.Println("Promotions already seeded, skipping")
		return
	}

	promotions := []promotion.Promotion{
		{
			ExternalID:  "prm-1",
			Title:       "Stay 3 Nights Pay 2",
			Description: "Book any room at The Phoenix Hotel Yogyakarta for 3 nights and only pay for 2 nights. Valid for deluxe and suite rooms.",
			Discount:    "33%",
			StartDate:   "2026-07-01",
			EndDate:     "2026-09-30",
			ImageURL:    "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600",
			Category:    "hotel",
			Status:      "active",
			Code:        "PHOENIX3FOR2",
		},
		{
			ExternalID:  "prm-2",
			Title:       "Jogja Food Trail 20% Off",
			Description: "Get 20% off on all food trail packages including Gudeg Yu Djum, Bakpia Pathok, and street food tours along Malioboro.",
			Discount:    "20%",
			StartDate:   "2026-07-01",
			EndDate:     "2026-12-31",
			ImageURL:    "https://images.unsplash.com/photo-1504674900247-0877df9cc836?q=80&w=600",
			Category:    "restaurant",
			Status:      "active",
			Code:        "JOGJAFOOD20",
		},
		{
			ExternalID:  "prm-3",
			Title:       "Merapi Lava Tour Bundle",
			Description: "Save 15% when you book a Merapi Jeep tour with a visit to Ullen Sentalu Museum and Selo Village lunch.",
			Discount:    "15%",
			StartDate:   "2026-08-01",
			EndDate:     "2026-10-31",
			ImageURL:    "https://images.unsplash.com/photo-1586319826484-b7f386bf3e72?auto=format&fit=crop&w=1200&q=80",
			Category:    "activity",
			Status:      "active",
			Code:        "MERAPI15",
		},
	}
	if err := db.CreateInBatches(promotions, 10).Error; err != nil {
		log.Printf("Failed to seed promotions: %v", err)
	}
	fmt.Printf("Promotions seeding done (%d records)\n", len(promotions))
}
